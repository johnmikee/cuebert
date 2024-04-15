package users

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type Query struct {
	config *Config
	sql    sq.SelectBuilder
}

func (c *Config) By() *Query {
	return &Query{
		config: c,
	}
}

// Query executes the query against the db with built query.
func (q *Query) Query() (UI, error) {
	sql, args, err := q.sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}
	users := []Info{}

	rows, err := q.config.db.Query(
		context.Background(),
		sql, args...)
	if err != nil {
		return nil, fmt.Errorf("user query failed %w", err)
	}
	for rows.Next() {
		var dev Info

		err = rows.Scan(
			&dev.UserServiceID,
			&dev.UserLongName,
			&dev.UserEmail,
			&dev.UserSlackID,
			&dev.TZOffset,
			&dev.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("user row query failed %w", err)
		}
		users = append(users, dev)
	}

	q.config.db.Release()

	return users, nil
}

func (q *Query) All() *Query {
	q.sql = q.config.st.Select(columns...).From(q.config.table)

	return q
}

func (q *Query) sqeq(column string, value interface{}) *Query {
	q.sql = q.config.st.Select(columns...).From(q.config.table).Where(sq.Eq{column: value})

	return q
}

func (q *Query) Opts(column Column, val any) *Query {
	return q.sqeq(string(column), val)
}
