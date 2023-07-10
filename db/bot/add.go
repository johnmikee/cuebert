package bot

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Update struct {
	args  []interface{}
	bresp Info
	db    *pgxpool.Conn
	query string

	st sq.StatementBuilderType

	ctx context.Context
	log logger.Logger
}

// AddBR initializes a new Update struct.
//
// the functions below that are methods of Update
// are used to modify specific fields of the statement that
// will be inserted once Execute is called.
func (c *Config) Add() *Update {
	return &Update{
		db:  c.db,
		ctx: context.Background(),
		log: c.log,
		st:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (c *Config) AddAllDevices(b []Info) (int64, error) {
	rows := [][]interface{}{}
	for i := range b {
		row := []interface{}{
			b[i].SlackID,
			b[i].UserEmail,
			b[i].ManagerSlackID,
			b[i].FirstACK,
			b[i].FirstACKTime,
			b[i].FirstMessageSent,
			b[i].FirstMessageSentAt,
			b[i].FirstMessageWaiting,
			b[i].ManagerMessageSent,
			b[i].ManagerMessageSentAt,
			b[i].FullName,
			b[i].DelayAt,
			b[i].DelayDate,
			b[i].DelayTime,
			b[i].DelaySent,
			b[i].ReminderInterval,
			b[i].ReminderWaiting,
			b[i].SerialNumber,
			b[i].TZOffset,
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

// Execute sends the statement to add the bot result after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *Update) Execute() error {
	query, args, err := u.st.Insert(table).
		Columns(columns...).
		Values(
			u.bresp.SlackID,
			u.bresp.UserEmail,
			u.bresp.ManagerSlackID,
			u.bresp.FirstACK,
			u.bresp.FirstACKTime,
			u.bresp.FirstMessageSent,
			u.bresp.FirstMessageSentAt,
			u.bresp.FirstMessageWaiting,
			u.bresp.ManagerMessageSent,
			u.bresp.ManagerMessageSentAt,
			u.bresp.FullName,
			u.bresp.DelayAt,
			u.bresp.DelayDate,
			u.bresp.DelayTime,
			u.bresp.DelaySent,
			u.bresp.ReminderInterval,
			u.bresp.ReminderWaiting,
			u.bresp.SerialNumber,
			u.bresp.TZOffset,
			helpers.UpdateTime(),
		).
		ToSql()

	if err != nil {
		return errors.Wrap(err, "failed to build sql statement")
	}

	_, err = u.db.Exec(
		u.ctx,
		query,
		args...)
	if err != nil {
		return err
	}

	u.log.Info().Msg("input was successfully submitted")
	u.log.Trace().
		Str("table", table).
		Interface("bot_response", u.bresp).
		Msg("inserted into table")

	u.db.Release()

	return nil
}

// Serial will update the value of the serial for the user response
func (u *Update) Serial(serial string) *Update {
	u.bresp.SerialNumber = serial

	return u
}

// SlackID will update the value of the users slack id
func (u *Update) SlackID(sid string) *Update {
	u.bresp.SlackID = sid

	return u
}

// FirstACK will update if the user acknowledged the first message
func (u *Update) FirstACK(f bool) *Update {
	u.bresp.FirstACK = f

	return u
}

// FirstMessageSent will update if the user was sent the first message
func (u *Update) FirstMessageSent(s bool) *Update {
	u.bresp.FirstMessageSent = s

	return u
}

// FirstMessageSentAt will update the time the first message was sent
func (u *Update) FirstMessageSentAt(t time.Time) *Update {
	u.bresp.FirstMessageSentAt = t

	return u
}

// FirstMessageWaiting will update if the user is waiting for the first message
func (u *Update) FirstMessageWaiting(w bool) *Update {
	u.bresp.FirstMessageWaiting = w

	return u
}

// FirstACKTime will update the time the user acknowledged the first message
func (u *Update) FirstACKTime(t time.Time) *Update {
	u.bresp.FirstACKTime = t

	return u
}

// FullName will update the users full name
func (u *Update) FullName(n string) *Update {
	u.bresp.FullName = n

	return u
}

// DelayAt will update the time the user delayed
func (u *Update) DelayAt(t time.Time) *Update {
	u.bresp.DelayAt = t

	return u
}

// DelayDate will update the date for reminder
func (u *Update) DelayDate(d string) *Update {
	u.bresp.DelayDate = d

	return u
}

// DelayTime will update the time for reminder
func (u *Update) DelayTime(d string) *Update {
	u.bresp.DelayTime = d

	return u
}

// DelaySent will update the status for the delay. If sent it will be true.
func (u *Update) DelaySent(d bool) *Update {
	u.bresp.DelaySent = d

	return u
}

// ManagerSlackID will update the value for the managers slack id
func (u *Update) ManagerID(m string) *Update {
	u.bresp.ManagerSlackID = m

	return u
}

// ManagerMessageSent will update the value for the manager message sent
func (u *Update) ManagerMessageSent(m bool) *Update {
	u.bresp.ManagerMessageSent = m

	return u
}

// ManagerMessageSentAt will update the value for the manager message sent at
func (u *Update) ManagerMessageSentAt(m time.Time) *Update {
	u.bresp.ManagerMessageSentAt = m

	return u
}

// ReminderInterval will update the value for the reminder interval
func (u *Update) ReminderInterval(r int) *Update {
	u.bresp.ReminderInterval = r

	return u
}

// ReminderWaiting will update the value for the reminder waiting
func (u *Update) ReminderWaiting(r bool) *Update {
	u.bresp.ReminderWaiting = r

	return u
}

// UserEmail will update the value for the users email
func (u *Update) UserEmail(e string) *Update {
	u.bresp.UserEmail = e

	return u
}
