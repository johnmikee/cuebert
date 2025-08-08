package tracking

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// Query represents the query builder for the tracking table.
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
func (q *Query) Query() (TI, error) {
	sql, args, err := q.sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}

	resp := []Info{}

	rows, err := q.config.db.Query(
		q.config.ctx,
		sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed %w", err)
	}

	for rows.Next() {
		var ti Info
		err = rows.Scan(
			&ti.SerialNumber,
			&ti.UserServiceID,
			&ti.TaskComplete,
			&ti.TaskName,
			&ti.UserSlackID,
			&ti.UserEmail,
			&ti.UserFullName,
			&ti.FirstACK,
			&ti.FirstACKTime,
			&ti.FirstMessageSent,
			&ti.FirstMessageSentAt,
			&ti.FirstMessageWaiting,
			&ti.UserReminderAt,
			&ti.UserReminderDate,
			&ti.UserReminderTime,
			&ti.ReminderSent,
			&ti.ReminderInterval,
			&ti.ReminderWaiting,
			&ti.TZOffset,
			&ti.UpdatedAt,
			&ti.FinalMessageSent,
			&ti.FinalMessageSentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("bot row query failed %w", err)
		}
		resp = append(resp, ti)
	}

	q.config.db.Release()

	return resp, nil
}

// WA is a where clause.
type WA struct {
	Column Column
	Value  any
}

// WhereAnd adds a where clause to the query.
func (q *Query) WhereAnd(w []WA) *Query {
	q.sql = q.config.st.Select(columns...).
		From(q.config.table).
		Where(
			sq.
				And{
				sq.Eq{string(w[0].Column): w[0].Value},
				sq.Eq{string(w[1].Column): w[1].Value},
			},
		)

	return q
}

// All returns all the information from the tracking table.
func (q *Query) All() *Query {
	q.sql = q.config.st.Select(columns...).From(q.config.table)

	return q
}

func (q *Query) sqeq(column string, value interface{}) *Query {
	q.sql = q.config.st.Select(columns...).From(q.config.table).Where(sq.Eq{column: value})

	return q
}

// Opts sets the options for the query.
func (q *Query) Opts(column Column, val any) *Query {
	return q.sqeq(string(column), val)
}
