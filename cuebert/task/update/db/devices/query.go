package devices

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type Query struct {
	config *Config
	sql    sq.SelectBuilder
}

// By returns a new client used to interact with specific columns
// in the devices table.
//
// Valid options are any of the columnds in the table which can be accessed
// by any of the methods of Query below.
func (c *Config) By() *Query {
	return &Query{
		config: c,
	}
}

// Query executes the query against the db with built query.
func (q *Query) Query() (DI, error) {
	sql, args, err := q.sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}
	q.config.log.Trace().Str("query", sql).Interface("args", args).Msg("composed sql query")

	devices := DI{}

	rows, err := q.config.db.Query(
		context.Background(),
		sql, args...)
	if err != nil {
		return nil, fmt.Errorf("device query failed %w", err)
	}

	for rows.Next() {
		var dev Info

		err = rows.Scan(
			&dev.DeviceID,
			&dev.DeviceName,
			&dev.Model,
			&dev.SerialNumber,
			&dev.Platform,
			&dev.OSVersion,
			&dev.User,
			&dev.UserMDMID,
			&dev.LastCheckIn,
			&dev.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("device row query failed %w", err)
		}
		devices = append(devices, dev)
	}

	q.config.db.Release()

	return devices, nil
}

// All returns all users and values in the table
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
