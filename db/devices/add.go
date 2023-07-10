package devices

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/pkg/errors"
)

type Update struct {
	args   []interface{}
	ctx    context.Context
	device *Info
	db     *pgxpool.Conn
	log    logger.Logger
	st     sq.StatementBuilderType
	query  string
}

// Add initializes a new Update struct.
//
// the functions below that are methods of Update
// are used to modify specific fields of the statement that
// will be inserted once Execute is called.
func (c *Config) Add() *Update {
	return &Update{
		db:     c.db,
		device: &Info{},
		ctx:    context.Background(),
		log:    c.log,
		st:     c.st,
	}
}

// AddAllDevices should only be used in specific cases.
//
// - to initialize the DB with values
// - if a table is dropped and rebuilt
//
// This will add all devices passed to the DB.
func (c *Config) AddAllDevices(devices DI) (int64, error) {
	rows := [][]interface{}{}
	for i := range devices {
		row := []interface{}{
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
			helpers.UpdateTime(),
		}
		rows = append(rows, row)
	}

	copyCount, err := c.db.CopyFrom(
		context.Background(),
		pgx.Identifier{table},
		columns,
		pgx.CopyFromRows(rows))

	c.db.Release()

	return copyCount, err
}

// Execute sends the statement to add the device after it has been composed.
func (u *Update) Execute() (*pgxpool.Conn, error) {
	query, args, err := u.st.Insert(table).
		Columns(columns...).
		Values(
			u.device.DeviceID,
			u.device.DeviceName,
			u.device.Model,
			u.device.SerialNumber,
			u.device.Platform,
			u.device.OSVersion,
			u.device.User,
			u.device.UserMDMID,
			u.device.LastCheckIn,
			helpers.UpdateTime(),
			helpers.UpdateTime(),
		).ToSql()

	if err != nil {
		return nil, errors.Wrap(err, "failed to build insert query")
	}
	_, err = u.db.Exec(
		u.ctx,
		query, args...)

	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	u.db.Release()
	u.log.Info().Msg("input was successfully submitted")

	return u.db, nil
}

// ID will update the value of the id for the device
func (u *Update) ID(id string) *Update {
	u.device.DeviceID = id

	return u
}

// Model will update the value of the model for the device.
func (u *Update) Model(model string) *Update {
	u.device.Model = model

	return u
}

// Name will update the value of the host name for the device.
func (u *Update) Name(name string) *Update {
	u.device.DeviceName = name

	return u
}

// OS will update the value of the reported OS for the device.
func (u *Update) OS(os string) *Update {
	u.device.OSVersion = os

	return u
}

// Platform will update the value of the platform for the device.
func (u *Update) Platform(platform string) *Update {
	u.device.Platform = platform

	return u
}

// Serial will update the value of the serial number for the device
func (u *Update) Serial(serial string) *Update {
	u.device.SerialNumber = serial

	return u
}

// User will update the value of the user for the device
func (u *Update) User(user string) *Update {
	u.device.User = user

	return u
}

// UserMDMID will update the value of the users id for the device
func (u *Update) UserMDMID(uid string) *Update {
	u.device.UserMDMID = uid

	return u
}
