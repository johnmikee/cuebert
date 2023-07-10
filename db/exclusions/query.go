package exclusions

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Query holds the configuration for the building and executing the query.
type Query struct {
	db  *pgxpool.Conn
	log logger.Logger
	sql sq.SelectBuilder
	st  sq.StatementBuilderType
}

// By returns a new client used to interact with specific columns
// in the exclusions table.
//
// Valid options are any of the columns in the table which can be accessed
// by any of the methods of Query below.
func (c *Config) Query() *Query {
	return &Query{
		db:  c.db,
		log: c.log,
		st:  c.st,
	}
}

// Query executes the query against the db with built query.
func (q *Query) Query() (EI, error) {
	sql, args, err := q.sql.ToSql()

	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}

	q.log.Trace().Str("query", sql).Msg("composed sql query")

	ex := []Info{}

	rows, err := q.db.Query(
		context.Background(),
		sql, args...)

	if err != nil {
		return nil, fmt.Errorf("exclusion query failed %w", err)
	}

	for rows.Next() {
		var e Info

		err = rows.Scan(
			&e.Approved,
			&e.SerialNumber,
			&e.UserEmail,
			&e.Reason,
			&e.Until,
			&e.CreatedAt,
			&e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("exclusion row query failed %w", err)
		}
		ex = append(ex, e)
	}

	q.db.Release()

	return ex, nil
}

// All returns all exclusions in the table
func (e *Query) All() *Query {
	e.sql = e.st.Select("*").From(table)

	return e
}

// Approved queries the exclusions table for approved requests.
func (e *Query) Approved(status bool) *Query {
	e.sql = e.st.Select("*").From(table).Where(sq.Eq{"approved": status})

	return e
}

// Serial queries the exclusions table for a specific serial number value
func (e *Query) Serial(id ...string) *Query {
	e.sql = e.st.Select("*").From(table).Where(sq.Eq{"serial_number": id})

	return e
}

// Email queries the exclusions table for a specific device id value
func (e *Query) Email(id ...string) *Query {
	e.sql = e.st.Select("*").From(table).Where(sq.Eq{"user_email": id})

	return e
}
