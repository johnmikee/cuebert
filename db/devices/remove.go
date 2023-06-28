package devices

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type DeviceRemove struct {
	ctx context.Context
	db  *pgxpool.Conn
	log logger.Logger
	sql sq.DeleteBuilder
	st  sq.StatementBuilderType
}

// Remove initializes a new DeviceRemove struct.
//
// the methods of DeviceRemove are used to designate specific
// fields of the statement that will be inserted once Execute is called.
func (c *Config) Remove() *DeviceRemove {
	return &DeviceRemove{
		ctx: c.ctx,
		db:  c.db,
		log: c.log,
		st:  c.st,
	}
}

// Remove sends the statement to remove the device after it has been composed.
func (u *DeviceRemove) Execute() (*pgxpool.Conn, error) {
	sql, args, err := u.sql.ToSql()

	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}
	u.log.Trace().Str("query", sql).Msg("composed sql query")

	_, err = u.db.Exec(
		u.ctx,
		sql, args...)

	if err != nil {
		return nil, fmt.Errorf("device query failed %w", err)
	}

	u.db.Release()
	u.log.Info().Msg("device was successfully removed")

	return u.db, nil
}

// ID will update the value of the id for the device
func (u *DeviceRemove) ID(id ...string) *DeviceRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"device_id": id})
	return u
}

// Model will update the value of the model for the device.
func (u *DeviceRemove) Model(model ...string) *DeviceRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"model": model})
	return u
}

// Name will update the value of the host name for the device.
func (u *DeviceRemove) Name(name ...string) *DeviceRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"device_name": name})
	return u
}

// OS will update the value of the reported OS for the device.
//   - this will very likely be the most used function.
func (u *DeviceRemove) OS(os ...string) *DeviceRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"os_version": os})
	return u
}

// Platform will update the value of the platform for the device.
func (u *DeviceRemove) Platform(platform ...string) *DeviceRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"platform": platform})
	return u
}

// Serial will update the value of the serial number for the device
func (u *DeviceRemove) Serial(serial ...string) *DeviceRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"serial_number": serial})
	return u
}

// User will update the value of the user for the device
func (u *DeviceRemove) User(user ...string) *DeviceRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_name": user})
	return u
}

// UserMDMID will update the value of the users id for the device
func (u *DeviceRemove) UserMDMID(uid string) *DeviceRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_mdm_id": uid})
	return u
}
