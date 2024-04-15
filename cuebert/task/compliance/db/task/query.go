package task

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// Query represents the query builder for the task table.
type Query struct {
	config *Config
	sql    sq.SelectBuilder
}

// By returns a new query builder.
func (c *Config) By() *Query {
	return &Query{
		config: c,
	}
}

// Query executes the query against the db with built query.
func (c *Query) Query() (TI, error) {
	sql, args, err := c.sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}

	resp := []Info{}

	rows, err := c.config.db.Query(
		c.config.ctx,
		sql, args...)
	if err != nil {
		return nil, fmt.Errorf("bot response query failed %w", err)
	}

	for rows.Next() {
		var ti Info
		err = rows.Scan(
			&ti.UserTaskID,
			&ti.UserID,
			&ti.TaskID,
			&ti.UserEmail,
			&ti.ItemName,
			&ti.TaskDescription,
			&ti.TaskComplete,
			&ti.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("bot row query failed %w", err)
		}
		resp = append(resp, ti)
	}

	c.config.db.Release()

	return resp, nil
}

// All returns all the tasks from the database.
func (q *Query) All() *Query {
	q.sql = q.config.st.Select(columns...).From(q.config.table)

	return q
}

func (q *Query) sqeq(column string, value interface{}) *Query {
	q.sql = q.config.st.Select(columns...).From(q.config.table).Where(sq.Eq{column: value})

	return q
}

// Opts returns the tasks from the database based on the options provided.
func (q *Query) Opts(column Column, val any) *Query {
	return q.sqeq(string(column), val)
}
