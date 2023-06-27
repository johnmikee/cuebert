package db

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/johnmikee/cuebert/db/devices"
)

type Overlap struct {
	Serial   string
	User     string
	SlackID  string
	OS       string
	TZOffset int64
	FullName string
}

type Devices = Config

func (d *Devices) add() *devices.DeviceUpdate {
	return d.dev(d.db, &d.log).Add()
}

// AddDevice provides access to the DeviceUpdate.Add() method
func (d *Devices) AddDevice() *devices.DeviceUpdate {
	return d.add()
}

func (d *Devices) remove() *devices.DeviceRemove {
	return d.dev(d.db, &d.log).Remove()
}

// RemoveDevice provides access to the DeviceRemove method
func (d *Devices) RemoveDeviceBy() *devices.DeviceRemove {
	return d.remove()
}

func (d *Devices) query() *devices.DeviceQuery {
	return d.dev(d.db, &d.log).By()
}

// QueryDeviceBy provides access to the DeviceQuery.By() method
func (d *Devices) QueryDeviceBy() *devices.DeviceQuery {
	return d.query()
}

func (d *Devices) update() *devices.DeviceUpdate {
	return d.dev(d.db, &d.log).Update()
}

// UpdateDeviceBy provides access to the DeviceUpdate.Update() method
func (d *Devices) UpdateDeviceBy() *devices.DeviceUpdate {
	return d.update()
}

// AddAll adds all devices provided to the devices table
func (d *Devices) AddAll(machines []devices.DeviceInfo) error {
	_, err := d.dev(d.db, &d.log).AddAllDevices(machines)

	return err
}

// GetAllDevices returns all devices from the devices table
func (d *Devices) GetAllDevices() ([]devices.DeviceInfo, error) {
	return d.query().All().Query()
}

// DeviceBySerial returns a device by serial number from the devices table
func (d *Devices) DeviceBySerial(serial string) (devices.DI, error) {
	return d.query().Serial(serial).Query()
}

// DeviceByEmail returns devices for a user from the devices table
func (d *Devices) DeviceByEmail(email string) ([]devices.DeviceInfo, error) {
	return d.query().User(email).Query()
}

// DevicesByUser returns all devices for a user from the devices table
func (d *Devices) DevicesByUser(user string) ([]devices.DeviceInfo, error) {
	userInfo, err := d.UserByID(user)
	if err != nil {
		return nil, err
	}

	users := []string{}
	for _, i := range userInfo {
		d.log.Info().Msg(i.MDMID)
		users = append(users, i.MDMID)
	}

	s, err := d.query().UserID(users[0]).Query()
	if err != nil {
		return s, err
	}

	return s, nil
}

// DeviceUserOverlap returns a list of devices that have a user name and a slack id
func (d *Devices) DeviceUserOverlap() ([]Overlap, error) {
	query, args, err := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("user_name, user_long_name, serial_number, user_slack_id, os_version, tz_offset").
		From("devices d").
		Join("users ON (user_email = d.user_name)").
		Where(sq.And{sq.NotEq{"user_name": ""}, sq.NotEq{"user_slack_id": ""}}).
		ToSql()

	d.log.Trace().Str("query", query).Interface("args", args).Msg("query")
	if err != nil {
		return nil, fmt.Errorf("device overlap query failed %w", err)
	}

	rows, err := d.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("bot response query failed %w", err)
	}

	resp := []Overlap{}
	for rows.Next() {
		var br Overlap

		err = rows.Scan(
			&br.User,
			&br.FullName,
			&br.Serial,
			&br.SlackID,
			&br.OS,
			&br.TZOffset,
		)

		if err != nil {
			return nil, fmt.Errorf("bot row query failed %w", err)
		}
		resp = append(resp, br)
	}

	return resp, nil
}
