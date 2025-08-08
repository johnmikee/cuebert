package common

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

// SafeQueries provides SQL-injection-safe database operations.
type SafeQueries struct {
	psql sq.StatementBuilderType
}

// NewSafeQueries creates a new instance with PostgreSQL placeholder format.
func NewSafeQueries() *SafeQueries {
	return &SafeQueries{
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// UpdateFirstMessageDelivered safely updates first message delivery status.
func (s *SafeQueries) UpdateFirstMessageDelivered(ctx context.Context, pool DatabasePool, tableName string, slackID string) error {
	query, args, err := s.psql.
		Update(tableName).
		Set("first_message_sent", true).
		Set("first_message_sent_at", helpers.UpdateTime()).
		Where(sq.Eq{"user_slack_id": slackID}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build first message update query")
	}

	_, err = pool.Exec(ctx, query, args...)
	if err != nil {
		return errors.DBError("update first message delivered", err)
	}

	return nil
}

// UpdateFirstMessageSent safely updates first message sent status.
func (s *SafeQueries) UpdateFirstMessageSent(ctx context.Context, pool DatabasePool, tableName string, slackID string) error {
	query, args, err := s.psql.
		Update(tableName).
		Set("first_message_sent", true).
		Set("first_message_sent_at", helpers.UpdateTime()).
		Where(sq.Eq{"user_slack_id": slackID}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build first message sent query")
	}

	_, err = pool.Exec(ctx, query, args...)
	if err != nil {
		return errors.DBError("update first message sent", err)
	}

	return nil
}

// UpdateFinalMessageSent safely updates final message sent status.
func (s *SafeQueries) UpdateFinalMessageSent(ctx context.Context, pool DatabasePool, tableName string, slackID string) error {
	query, args, err := s.psql.
		Update(tableName).
		Set("final_message_sent", true).
		Set("final_message_sent_at", helpers.UpdateTime()).
		Where(sq.Eq{"user_slack_id": slackID}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build final message sent query")
	}

	_, err = pool.Exec(ctx, query, args...)
	if err != nil {
		return errors.DBError("update final message sent", err)
	}

	return nil
}

// DeleteTrackingRecords safely deletes tracking records by serial numbers.
func (s *SafeQueries) DeleteTrackingRecords(ctx context.Context, pool DatabasePool, tableName string, serialNumbers []string) (int64, error) {
	if len(serialNumbers) == 0 {
		return 0, nil
	}

	query, args, err := s.psql.
		Delete(tableName).
		Where(sq.Eq{"serial_number": serialNumbers}).
		ToSql()
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build delete query")
	}

	result, err := pool.Exec(ctx, query, args...)
	if err != nil {
		return 0, errors.DBError("delete tracking records", err)
	}

	return result.RowsAffected(), nil
}

// UpdateReminderSent safely updates reminder sent status.
func (s *SafeQueries) UpdateReminderSent(ctx context.Context, pool DatabasePool, tableName string, slackID string) error {
	query, args, err := s.psql.
		Update(tableName).
		Set("reminder_sent", true).
		Set("reminder_sent_at", time.Now()).
		Set("reminder_waiting", false).
		Where(sq.Eq{"user_slack_id": slackID}).
		ToSql()
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeDBQuery, "failed to build reminder sent query")
	}

	_, err = pool.Exec(ctx, query, args...)
	if err != nil {
		return errors.DBError("update reminder sent", err)
	}

	return nil
}

// DatabasePool interface for database operations.
type DatabasePool interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (Result, error)
}

// Result interface for database command results.
type Result interface {
	RowsAffected() int64
}

// PgxPoolAdapter adapts pgxpool.Pool to DatabasePool interface.
type PgxPoolAdapter struct {
	pool *pgxpool.Pool
}

// NewPgxPoolAdapter creates a new adapter for pgxpool.Pool.
func NewPgxPoolAdapter(pool *pgxpool.Pool) *PgxPoolAdapter {
	return &PgxPoolAdapter{pool: pool}
}

// Exec implements DatabasePool interface.
func (a *PgxPoolAdapter) Exec(ctx context.Context, sql string, arguments ...interface{}) (Result, error) {
	tag, err := a.pool.Exec(ctx, sql, arguments...)
	if err != nil {
		return nil, err
	}
	return &pgxResult{tag: tag}, nil
}

// pgxResult adapts pgconn.CommandTag to Result interface.
type pgxResult struct {
	tag pgconn.CommandTag
}

// RowsAffected implements Result interface.
func (r *pgxResult) RowsAffected() int64 {
	return r.tag.RowsAffected()
}
