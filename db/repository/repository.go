package repository

import (
	"context"
	"fmt"
	"reflect"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Repository defines the interface for generic database operations.
type Repository[T any] interface {
	// Create operations
	Create(ctx context.Context, item T) error
	CreateBatch(ctx context.Context, items []T) error

	// Read operations
	FindByID(ctx context.Context, id interface{}) (*T, error)
	FindOne(ctx context.Context, conditions Conditions) (*T, error)
	FindAll(ctx context.Context, conditions Conditions) ([]T, error)
	Exists(ctx context.Context, conditions Conditions) (bool, error)
	Count(ctx context.Context, conditions Conditions) (int64, error)

	// Update operations
	Update(ctx context.Context, id interface{}, updates Updates) error
	UpdateWhere(ctx context.Context, conditions Conditions, updates Updates) (int64, error)

	// Delete operations
	Delete(ctx context.Context, id interface{}) error
	DeleteWhere(ctx context.Context, conditions Conditions) (int64, error)

	// Transaction support
	WithTx(tx pgx.Tx) Repository[T]
}

// Conditions represents query conditions.
type Conditions map[string]interface{}

// Updates represents update values.
type Updates map[string]interface{}

// OrderBy represents ordering.
type OrderBy struct {
	Column string
	Desc   bool
}

// QueryOptions represents additional query options.
type QueryOptions struct {
	OrderBy []OrderBy
	Limit   int
	Offset  int
}

// BaseRepository provides a generic implementation of Repository.
type BaseRepository[T any] struct {
	db         *pgxpool.Pool
	tx         pgx.Tx
	log        logger.Logger
	table      string
	primaryKey string
	columns    []string
	psql       sq.StatementBuilderType
}

// Config holds repository configuration.
type Config struct {
	DB         *pgxpool.Pool
	Log        logger.Logger
	Table      string
	PrimaryKey string
}

// New creates a new repository instance.
func New[T any](cfg Config) Repository[T] {
	repo := &BaseRepository[T]{
		db:         cfg.DB,
		log:        logger.ChildLogger(fmt.Sprintf("repo_%s", cfg.Table), &cfg.Log),
		table:      cfg.Table,
		primaryKey: cfg.PrimaryKey,
		psql:       sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}

	// Extract column names from struct tags
	repo.columns = extractColumns[T]()

	return repo
}

// Create inserts a new record.
func (r *BaseRepository[T]) Create(ctx context.Context, item T) error {
	values := structToMap(item)
	values["updated_at"] = helpers.UpdateTime()

	query := r.psql.Insert(r.table).
		SetMap(values).
		Suffix("ON CONFLICT DO NOTHING")

	sql, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build insert query")
	}

	if r.tx != nil {
		_, err = r.tx.Exec(ctx, sql, args...)
	} else {
		_, err = r.db.Exec(ctx, sql, args...)
	}

	if err != nil {
		return errors.DBError("create record", err)
	}

	return nil
}

// CreateBatch inserts multiple records efficiently.
func (r *BaseRepository[T]) CreateBatch(ctx context.Context, items []T) error {
	if len(items) == 0 {
		return nil
	}

	// Start a transaction for batch insert
	conn, err := r.getConn(ctx)
	if err != nil {
		return err
	}

	if r.tx == nil {
		defer conn.Release()
	}

	batch := &pgx.Batch{}

	for _, item := range items {
		values := structToMap(item)
		values["updated_at"] = helpers.UpdateTime()

		query := r.psql.Insert(r.table).
			SetMap(values).
			Suffix("ON CONFLICT DO NOTHING")

		sql, args, err := query.ToSql()
		if err != nil {
			return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build batch insert query")
		}

		batch.Queue(sql, args...)
	}

	var results pgx.BatchResults
	if r.tx != nil {
		results = r.tx.SendBatch(ctx, batch)
	} else {
		results = conn.SendBatch(ctx, batch)
	}
	defer results.Close()

	for range items {
		_, err := results.Exec()
		if err != nil {
			return errors.DBError("batch create", err)
		}
	}

	return nil
}

// FindByID retrieves a record by primary key.
func (r *BaseRepository[T]) FindByID(ctx context.Context, id interface{}) (*T, error) {
	return r.FindOne(ctx, Conditions{r.primaryKey: id})
}

// FindOne retrieves a single record matching conditions.
func (r *BaseRepository[T]) FindOne(ctx context.Context, conditions Conditions) (*T, error) {
	query := r.psql.Select(r.columns...).
		From(r.table).
		Limit(1)

	query = applyConditions(query, conditions)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build select query")
	}

	var row pgx.Row
	if r.tx != nil {
		row = r.tx.QueryRow(ctx, sql, args...)
	} else {
		row = r.db.QueryRow(ctx, sql, args...)
	}

	var result T
	err = scanRow(row, &result)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.NotFoundError(r.table)
		}
		return nil, errors.DBError("find one", err)
	}

	return &result, nil
}

// FindAll retrieves all records matching conditions.
func (r *BaseRepository[T]) FindAll(ctx context.Context, conditions Conditions) ([]T, error) {
	query := r.psql.Select(r.columns...).From(r.table)
	query = applyConditions(query, conditions)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build select query")
	}

	var rows pgx.Rows
	if r.tx != nil {
		rows, err = r.tx.Query(ctx, sql, args...)
	} else {
		rows, err = r.db.Query(ctx, sql, args...)
	}
	if err != nil {
		return nil, errors.DBError("find all", err)
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		var item T
		err := scanRows(rows, &item)
		if err != nil {
			return nil, errors.DBError("scan row", err)
		}
		results = append(results, item)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.DBError("iterate rows", err)
	}

	return results, nil
}

// Exists checks if any records match conditions.
func (r *BaseRepository[T]) Exists(ctx context.Context, conditions Conditions) (bool, error) {
	query := r.psql.Select("1").
		From(r.table).
		Limit(1)

	query = applyConditions(query, conditions)

	sql, args, err := query.ToSql()
	if err != nil {
		return false, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build exists query")
	}

	var exists int
	var row pgx.Row
	if r.tx != nil {
		row = r.tx.QueryRow(ctx, sql, args...)
	} else {
		row = r.db.QueryRow(ctx, sql, args...)
	}

	err = row.Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, errors.DBError("check exists", err)
	}

	return true, nil
}

// Count returns the number of records matching conditions.
func (r *BaseRepository[T]) Count(ctx context.Context, conditions Conditions) (int64, error) {
	query := r.psql.Select("COUNT(*)").From(r.table)
	query = applyConditions(query, conditions)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build count query")
	}

	var count int64
	var row pgx.Row
	if r.tx != nil {
		row = r.tx.QueryRow(ctx, sql, args...)
	} else {
		row = r.db.QueryRow(ctx, sql, args...)
	}

	err = row.Scan(&count)
	if err != nil {
		return 0, errors.DBError("count", err)
	}

	return count, nil
}

// Update modifies a record by ID.
func (r *BaseRepository[T]) Update(ctx context.Context, id interface{}, updates Updates) error {
	updates["updated_at"] = helpers.UpdateTime()

	query := r.psql.Update(r.table).
		SetMap(updates).
		Where(sq.Eq{r.primaryKey: id})

	sql, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build update query")
	}

	var result pgconn.CommandTag
	if r.tx != nil {
		result, err = r.tx.Exec(ctx, sql, args...)
	} else {
		result, err = r.db.Exec(ctx, sql, args...)
	}

	if err != nil {
		return errors.DBError("update", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NotFoundError(r.table)
	}

	return nil
}

// UpdateWhere modifies records matching conditions.
func (r *BaseRepository[T]) UpdateWhere(ctx context.Context, conditions Conditions, updates Updates) (int64, error) {
	updates["updated_at"] = helpers.UpdateTime()

	query := r.psql.Update(r.table).SetMap(updates)

	for key, value := range conditions {
		query = query.Where(sq.Eq{key: value})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build update query")
	}

	var result pgconn.CommandTag
	if r.tx != nil {
		result, err = r.tx.Exec(ctx, sql, args...)
	} else {
		result, err = r.db.Exec(ctx, sql, args...)
	}

	if err != nil {
		return 0, errors.DBError("update where", err)
	}

	return result.RowsAffected(), nil
}

// Delete removes a record by ID.
func (r *BaseRepository[T]) Delete(ctx context.Context, id interface{}) error {
	query := r.psql.Delete(r.table).Where(sq.Eq{r.primaryKey: id})

	sql, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build delete query")
	}

	var result pgconn.CommandTag
	if r.tx != nil {
		result, err = r.tx.Exec(ctx, sql, args...)
	} else {
		result, err = r.db.Exec(ctx, sql, args...)
	}

	if err != nil {
		return errors.DBError("delete", err)
	}

	if result.RowsAffected() == 0 {
		return errors.NotFoundError(r.table)
	}

	return nil
}

// DeleteWhere removes records matching conditions.
func (r *BaseRepository[T]) DeleteWhere(ctx context.Context, conditions Conditions) (int64, error) {
	query := r.psql.Delete(r.table)

	for key, value := range conditions {
		query = query.Where(sq.Eq{key: value})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build delete query")
	}

	var result pgconn.CommandTag
	if r.tx != nil {
		result, err = r.tx.Exec(ctx, sql, args...)
	} else {
		result, err = r.db.Exec(ctx, sql, args...)
	}

	if err != nil {
		return 0, errors.DBError("delete where", err)
	}

	return result.RowsAffected(), nil
}

// WithTx returns a repository that uses the given transaction.
func (r *BaseRepository[T]) WithTx(tx pgx.Tx) Repository[T] {
	return &BaseRepository[T]{
		db:         r.db,
		tx:         tx,
		log:        r.log,
		table:      r.table,
		primaryKey: r.primaryKey,
		columns:    r.columns,
		psql:       r.psql,
	}
}

// getConn gets a connection from pool or returns existing transaction.
func (r *BaseRepository[T]) getConn(ctx context.Context) (*pgxpool.Conn, error) {
	if r.tx != nil {
		return nil, nil // Using transaction
	}

	conn, err := r.db.Acquire(ctx)
	if err != nil {
		return nil, errors.DBError("acquire connection", err)
	}

	return conn, nil
}

// Helper functions

func applyConditions(query sq.SelectBuilder, conditions Conditions) sq.SelectBuilder {
	for key, value := range conditions {
		if value == nil {
			query = query.Where(sq.Eq{key: nil})
		} else if v, ok := value.([]interface{}); ok {
			query = query.Where(sq.Eq{key: v})
		} else {
			query = query.Where(sq.Eq{key: value})
		}
	}
	return query
}

func extractColumns[T any]() []string {
	var item T
	t := reflect.TypeOf(item)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		if tag != "" && tag != "-" {
			columns = append(columns, tag)
		}
	}

	return columns
}

func structToMap(item interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(item)
	t := reflect.TypeOf(item)

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")

		if tag != "" && tag != "-" {
			value := v.Field(i).Interface()

			// Handle zero time values
			if timeVal, ok := value.(time.Time); ok && timeVal.IsZero() {
				continue
			}

			result[tag] = value
		}
	}

	return result
}

// scanRow and scanRows need to be implemented based on your struct scanning needs
// These are placeholder implementations.
func scanRow(row pgx.Row, dest interface{}) error {
	// This would use reflection or a scanning library to map columns to struct fields
	// For now, delegate to the row's Scan method
	return row.Scan(dest)
}

func scanRows(rows pgx.Rows, dest interface{}) error {
	// This would use reflection or a scanning library to map columns to struct fields
	// For now, delegate to the rows' Scan method
	return rows.Scan(dest)
}
