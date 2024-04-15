package devices

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/pkg/errors"
)

type Add struct {
	di     Info
	config *Config
}

// Add initializes a new Update struct.
//
// the functions below that are methods of Update
// are used to modify specific fields of the statement that
// will be inserted once Execute is called.
func (c *Config) Add() *Add {
	return &Add{
		config: c,
	}
}

// AddAllDevices should only be used in specific cases.
//
// - to initialize the DB with values
// - if a table is dropped and rebuilt
//
// This will add all devices passed to the DB.
func (c *Config) AddAllDevices(devices DI) (int64, error) {
	for i := range devices {
		query, args, err := c.st.Insert(c.table).Columns(columns...).
			Values(
				devices[i].DeviceID,
				devices[i].DeviceName,
				devices[i].Model,
				devices[i].SerialNumber,
				devices[i].Platform,
				devices[i].OSVersion,
				devices[i].User,
				devices[i].UserMDMID,
				devices[i].LastCheckIn,
				helpers.UpdateTime(),
			).Suffix("ON CONFLICT DO NOTHING").ToSql()
		if err != nil {
			c.log.Debug().AnErr("error", err).Msg("could not build query")
			continue
		}

		_, err = c.db.Exec(context.Background(), query, args...)
		if err != nil {
			c.log.Debug().AnErr("error", err).Msg("could not execute query")
			continue
		}

	}

	return 0, nil
}

// Execute sends the statement to add the device after it has been composed.
func (a *Add) Execute() (*pgxpool.Conn, error) {
	query, args, err := a.config.st.Insert(a.config.table).
		Columns(columns...).
		Values(
			a.di.DeviceID,
			a.di.DeviceName,
			a.di.Model,
			a.di.SerialNumber,
			a.di.Platform,
			a.di.OSVersion,
			a.di.User,
			a.di.UserMDMID,
			a.di.LastCheckIn,
			helpers.UpdateTime(),
		).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build insert query")
	}
	_, err = a.config.db.Exec(
		a.config.ctx,
		query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	a.config.db.Release()
	a.config.log.Info().Msg("input was successfully submitted")

	return a.config.db, nil
}
