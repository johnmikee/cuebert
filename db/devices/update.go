package devices

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/parser"
)

// Update initializes a new Update struct.
//
// The methods of Update are used to modify which values
// will be updated. those are parsed and the
// column we are using as the condition to match is passed as the index.
func (c *Config) Update() *Update {
	return &Update{
		db:     c.db,
		device: &Info{},
		ctx:    context.Background(),
		log:    c.log,
		st:     c.st,
	}
}

// Send sends the statement to update the device after it has been composed.
func (u *Update) Send() (*pgxpool.Conn, error) {
	_, err := u.db.Exec(
		u.ctx,
		u.query, u.args...)

	if err != nil {
		return nil, err
	}

	u.db.Release()
	u.log.Info().Msg("input was successfully submitted")

	return u.db, nil
}

// Parse will take the input provided by the user via the methods
// of Update and compose a statement.
func (u *Update) Parse(index, val string) *Update {
	check := []parser.CheckInfo{
		{
			Fn: parser.Prim{
				S: u.device.DeviceName,
			},
			Key:     "device_name",
			Trimmed: "DeviceName",
		},
		{
			Fn: parser.Prim{
				T: *u.device.LastCheckIn,
			},
			Key:     "last_check_in",
			Trimmed: "LastCheckIn",
		},
		{
			Fn: parser.Prim{
				S: u.device.Model,
			},
			Key:     "model",
			Trimmed: "Model",
		},
		{
			Fn: parser.Prim{
				S: u.device.OSVersion,
			},
			Key:     "os_version",
			Trimmed: "OSVersion",
		},
		{
			Fn: parser.Prim{
				S: u.device.Platform,
			},
			Key:     "platform",
			Trimmed: "Platform",
		},
		{
			Fn: parser.Prim{
				S: u.device.SerialNumber,
			},
			Key:     "serial_number",
			Trimmed: "SerialNumber",
		},
		{
			Fn: parser.Prim{
				S: u.device.User,
			},
			Key:     "user",
			Trimmed: "User",
		},
		{
			Fn: parser.Prim{
				S: u.device.UserMDMID,
			},
			Key:     "user_mdm_id",
			Trimmed: "UserMDMID",
		},
	}

	query, args, err := parser.ParseInput(&parser.Parser{
		Index:  index,
		Table:  table,
		Val:    val,
		Check:  check,
		Method: parser.Update,
		Into:   Info{},
	})
	if err != nil {
		u.log.Err(err).Msg("error with parse input")
		return nil
	}
	u.args = args
	u.query = query

	return u
}
