package exclusions

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// ExclusionQuery holds the configuration for the building and executing the query.
type ExclusionQuery struct {
	db  *pgxpool.Conn
	log logger.Logger
	sql sq.SelectBuilder
	st  sq.StatementBuilderType
}

// By returns a new client used to interact with specific columns
// in the exclusions table.
//
// Valid options are any of the columns in the table which can be accessed
// by any of the methods of ExclusionQuery below.
func (c *Config) By() *ExclusionQuery {
	return &ExclusionQuery{
		db:  c.db,
		log: c.log,
		st:  c.st,
	}
}

// Query executes the query against the db with built query.
func (q *ExclusionQuery) Query() (EI, error) {
	sql, args, err := q.sql.ToSql()

	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}

	q.log.Trace().Str("query", sql).Msg("composed sql query")

	ex := []ExclusionInfo{}

	rows, err := q.db.Query(
		context.Background(),
		sql, args...)

	if err != nil {
		return nil, fmt.Errorf("exclusion query failed %w", err)
	}

	for rows.Next() {
		var e ExclusionInfo

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
func (e *ExclusionQuery) All() *ExclusionQuery {
	e.sql = e.st.Select("*").From(table)

	return e
}

// Approved queries the exclusions table for approved requests.
func (e *ExclusionQuery) Approved(status bool) *ExclusionQuery {
	e.sql = e.st.Select("*").From(table).Where(sq.Eq{"approved": status})

	return e
}

// Serial queries the exclusions table for a specific serial number value
func (e *ExclusionQuery) Serial(id ...string) *ExclusionQuery {
	e.sql = e.st.Select("*").From(table).Where(sq.Eq{"serial_number": id})

	return e
}

// Email queries the exclusions table for a specific device id value
func (e *ExclusionQuery) Email(id ...string) *ExclusionQuery {
	e.sql = e.st.Select("*").From(table).Where(sq.Eq{"user_email": id})

	return e
}
