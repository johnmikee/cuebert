package users

import (
	"context"
	"fmt"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Query struct {
	sql sq.SelectBuilder
	st  sq.StatementBuilderType
	db  *pgxpool.Conn
	log logger.Logger
}

// UserBy returns a new client used to interact with specific columns
// in the users table.
//
// Valid options are any of the columns in the table which can be accessed
// by any of the methods of Query below.
func (c *Config) By() *Query {
	return &Query{
		db:  c.db,
		log: c.log,
		st:  c.st,
	}
}

// Query executes the query against the db with built query.
func (q *Query) Query() ([]Info, error) {
	sql, args, err := q.sql.ToSql()

	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}
	q.log.Trace().Str("query", sql).Interface("args", args).Msg("composed sql query")

	users := []Info{}

	rows, err := q.db.Query(
		context.Background(),
		sql, args...)

	if err != nil {
		return nil, fmt.Errorf("user query failed %w", err)
	}
	for rows.Next() {
		var dev Info

		err = rows.Scan(
			&dev.MDMID,
			&dev.UserLongName,
			&dev.UserEmail,
			&dev.UserSlackID,
			&dev.TZOffset,
			&dev.CreatedAt,
			&dev.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("user row query failed %w", err)
		}
		users = append(users, dev)
	}

	q.db.Release()

	return users, nil
}

// All returns all users and values in the table
func (q *Query) All() *Query {
	q.sql = q.st.Select("*").From(table)

	return q
}

// ID queries the users table for a specific device id value
func (q *Query) ID(id ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"user_mdm_id": id})

	return q
}

// Created queries the users table for a specific created_at value
func (q *Query) Created(created string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"created_at": created})

	return q
}

// Email queries the users table for a specific users email
func (q *Query) Email(email ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"user_email": email})

	return q
}

// LongName queries the users table for a users long name
func (q *Query) LongName(name ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"user_long_name": name})

	return q
}

// SlackID queries the users table for a users slack_id
func (q *Query) SlackID(s ...string) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"user_slack_id": s})

	return q
}

// TZ queries the users table for a given tz_offset
func (q *Query) TZ(t int64) *Query {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"tz_offset": t})

	return q
}

// TZs queries the users table for multiple tz_offsets
func (q *Query) TZs(tzs []int64) *Query {
	s := []string{}
	for _, i := range tzs {
		s = append(s, strconv.Itoa(int(i)))
	}

	q.sql = q.st.Select("*").From(table).Where("tz_offset", s)

	return q
}
