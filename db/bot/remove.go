package bot

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Remove struct {
	sql sq.DeleteBuilder
	st  sq.StatementBuilderType

	db  *pgxpool.Conn
	ctx context.Context
	log logger.Logger
}

// Remove initializes a new Remove struct.
//
// the methods of Remove are used to designate specific
// fields of the statement that will be inserted once Execute is called.
func (c *Config) Remove() *Remove {
	return &Remove{
		db:  c.db,
		ctx: context.Background(),
		log: c.log,
		st:  c.st,
	}
}

// Execute sends the statement to remove the device after it has been composed.
func (u *Remove) Execute() (*pgxpool.Conn, error) {
	sql, args, err := u.sql.ToSql()
	if err != nil {
		return nil, err
	}
	u.log.Trace().Str("query", sql).Interface("args", args).Msg("query")
	_, err = u.db.Exec(
		u.ctx,
		sql, args...)

	if err != nil {
		return nil, err
	}

	u.db.Release()
	u.log.Info().Msg("result was successfully removed")

	return u.db, nil
}

// All will remove all results
func (u *Remove) All() *Remove {
	u.sql = u.st.Delete(table)

	return u
}

// DelaySent will remove based off if the delay has been sent or not
func (u *Remove) DelaySent(d bool) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"delay_sent": d})

	return u
}

// FirstMessageWaiting will remove based off if the first message is waiting or not
func (u *Remove) FirstMessageWaiting(d bool) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"first_message_waiting": d})

	return u
}

// FullName will remove based off of the full name for the device
func (u *Remove) FullName(name ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"full_name": name})

	return u
}

// ManagerSlackID will remove based off slack id
func (u *Remove) ManagerSlackID(id ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"manager_slack_id": id})

	return u
}

// ManagerMessageSent will remove based off if the manager message has been sent or not
func (u *Remove) ManagerMessageSent(d bool) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"manager_message_sent": d})

	return u
}

// SlackID will remove based off slack id
func (u *Remove) SlackID(id ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"slack_id": id})

	return u
}

// Serial will will remove based off of the serial number for the device
func (u *Remove) Serial(serial ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"serial_number": serial})

	return u
}

// UserEmail will remove based off of the user email for the device
func (u *Remove) UserEmail(email ...string) *Remove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_email": email})

	return u
}
