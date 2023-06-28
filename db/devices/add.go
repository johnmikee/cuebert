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

type DeviceUpdate struct {
	args   []interface{}
	ctx    context.Context
	device *DeviceInfo
	db     *pgxpool.Conn
	log    logger.Logger
	st     sq.StatementBuilderType
	query  string
}

// Add initializes a new DeviceUpdate struct.
//
// the functions below that are methods of DeviceUpdate
// are used to modify specific fields of the statement that
// will be inserted once Execute is called.
func (c *Config) Add() *DeviceUpdate {
	return &DeviceUpdate{
		db:     c.db,
		device: &DeviceInfo{},
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
func (c *Config) AddAllDevices(devices []DeviceInfo) (int64, error) {
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
func (u *DeviceUpdate) Execute() (*pgxpool.Conn, error) {
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
func (u *DeviceUpdate) ID(id string) *DeviceUpdate {
	u.device.DeviceID = id

	return u
}

// Model will update the value of the model for the device.
func (u *DeviceUpdate) Model(model string) *DeviceUpdate {
	u.device.Model = model

	return u
}

// Name will update the value of the host name for the device.
func (u *DeviceUpdate) Name(name string) *DeviceUpdate {
	u.device.DeviceName = name

	return u
}

// OS will update the value of the reported OS for the device.
func (u *DeviceUpdate) OS(os string) *DeviceUpdate {
	u.device.OSVersion = os

	return u
}

// Platform will update the value of the platform for the device.
func (u *DeviceUpdate) Platform(platform string) *DeviceUpdate {
	u.device.Platform = platform

	return u
}

// Serial will update the value of the serial number for the device
func (u *DeviceUpdate) Serial(serial string) *DeviceUpdate {
	u.device.SerialNumber = serial

	return u
}

// User will update the value of the user for the device
func (u *DeviceUpdate) User(user string) *DeviceUpdate {
	u.device.User = user

	return u
}

// UserMDMID will update the value of the users id for the device
func (u *DeviceUpdate) UserMDMID(uid string) *DeviceUpdate {
	u.device.UserMDMID = uid

	return u
}
