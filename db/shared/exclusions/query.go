package exclusions

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// Query holds the configuration for the building and executing the query.
type Query struct {
	config *Config
	sql    sq.SelectBuilder
}

// By returns a new client used to interact with specific columns
// in the exclusions table.
//
// Valid options are any of the columns in the table which can be accessed
// by any of the methods of Query below.
func (c *Config) By() *Query {
	return &Query{
		config: c,
	}
}

// Query executes the query against the db with built query.
func (q *Query) Query() (EI, error) {
	sql, args, err := q.sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}

	q.config.log.Trace().Str("query", sql).Msg("composed sql query")

	ex := []Info{}

	rows, err := q.config.db.Query(
		context.Background(),
		sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exclusion query failed %w", err)
	}

	for rows.Next() {
		var e Info

		err = rows.Scan(
			&e.Approved,
			&e.UserEmail,
			&e.Reason,
			&e.Until,
			&e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("exclusion row query failed %w", err)
		}
		ex = append(ex, e)
	}

	q.config.db.Release()

	return ex, nil
}

// All returns all exclusions in the table.
func (q *Query) All() *Query {
	q.sql = q.config.st.Select(q.config.columns...).From(q.config.table)

	return q
}

func (q *Query) sqeq(column string, value interface{}) *Query {
	q.sql = q.config.st.Select(q.config.columns...).From(q.config.table).Where(sq.Eq{column: value})

	return q
}

func (q *Query) Opts(column Column, val any) *Query {
	return q.sqeq(string(column), val)
}
