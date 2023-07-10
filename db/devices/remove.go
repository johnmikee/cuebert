package devices

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Remove struct {
	ctx context.Context
	db  *pgxpool.Conn
	log logger.Logger
	sql sq.DeleteBuilder
	st  sq.StatementBuilderType
}

// Remove initializes a new Remove struct.
//
// the methods of Remove are used to designate specific
// fields of the statement that will be inserted once Execute is called.
func (c *Config) Remove() *Remove {
	return &Remove{
		ctx: c.ctx,
		db:  c.db,
		log: c.log,
		st:  c.st,
	}
}

// Remove sends the statement to remove the device after it has been composed.
func (u *Remove) Execute() (*pgxpool.Conn, error) {
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
func (u *Remove) ID(id ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"device_id": id})
	return u
}

// Model will update the value of the model for the device.
func (u *Remove) Model(model ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"model": model})
	return u
}

// Name will update the value of the host name for the device.
func (u *Remove) Name(name ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"device_name": name})
	return u
}

// OS will update the value of the reported OS for the device.
//   - this will very likely be the most used function.
func (u *Remove) OS(os ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"os_version": os})
	return u
}

// Platform will update the value of the platform for the device.
func (u *Remove) Platform(platform ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"platform": platform})
	return u
}

// Serial will update the value of the serial number for the device
func (u *Remove) Serial(serial ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"serial_number": serial})
	return u
}

// User will update the value of the user for the device
func (u *Remove) User(user ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_name": user})
	return u
}

// UserMDMID will update the value of the users id for the device
func (u *Remove) UserMDMID(uid string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_mdm_id": uid})
	return u
}
