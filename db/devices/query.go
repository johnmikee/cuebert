package devices

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type DeviceQuery struct {
	db  *pgxpool.Conn
	log logger.Logger
	sql sq.SelectBuilder
	st  sq.StatementBuilderType
}

// By returns a new client used to interact with specific columns
// in the devices table.
//
// Valid options are any of the columnds in the table which can be accessed
// by any of the methods of DeviceQuery below.
func (c *Config) By() *DeviceQuery {
	return &DeviceQuery{
		db:  c.db,
		log: c.log,
		st:  c.st,
	}
}

// Query executes the query against the db with built query.
func (q *DeviceQuery) Query() (DI, error) {
	sql, args, err := q.sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}
	q.log.Trace().Str("query", sql).Interface("args", args).Msg("composed sql query")

	devices := []DeviceInfo{}

	rows, err := q.db.Query(
		context.Background(),
		sql, args...)
	if err != nil {
		return nil, fmt.Errorf("device query failed %w", err)
	}

	for rows.Next() {
		var dev DeviceInfo

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
			&dev.CreatedAt,
			&dev.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("device row query failed %w", err)
		}
		devices = append(devices, dev)
	}

	q.db.Release()

	return devices, nil
}

// All returns all devices and values in the table
func (q *DeviceQuery) All() *DeviceQuery {
	q.sql = q.st.Select("*").From(table)
	return q
}

// ID queries the devices table for a specific device id value
func (q *DeviceQuery) ID(id ...string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"device_id": id})

	return q
}

// Created queries the devices table for a specific created_at value
func (q *DeviceQuery) Created(created string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"created_at": created})

	return q
}

// Model queries the devices table for a specific model value
func (q *DeviceQuery) Model(model ...string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"model": model})

	return q
}

// Name queries the devices table for a specific host name
func (q *DeviceQuery) Name(model ...string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"device_name": model})

	return q
}

// OS queries the devices table for a specific os_version value
func (q *DeviceQuery) OS(os ...string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"os_version": os})

	return q
}

// Platform queries the devices table for a specific platform value
func (q *DeviceQuery) Platform(platform ...string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"platform": platform})

	return q
}

// Serial queries the devices table for a specific serial_number value
func (q *DeviceQuery) Serial(serial ...string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"serial_number": serial})
	return q
}

// User queries the devices table for a specific user value
func (q *DeviceQuery) User(user ...string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"user_name": user})

	return q
}

// UserID queries the devices table for a specific user id value
func (q *DeviceQuery) UserID(user ...string) *DeviceQuery {
	q.sql = q.st.Select("*").From(table).Where(sq.Eq{"user_mdm_id": user})

	return q
}
