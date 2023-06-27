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

type BotResUpdate struct {
	args  []interface{}
	bresp BotResInfo
	db    *pgxpool.Conn
	query string

	st sq.StatementBuilderType

	ctx context.Context
	log logger.Logger
}

// AddBR initializes a new BotResUpdate struct.
//
// the functions below that are methods of BotResUpdate
// are used to modify specific fields of the statement that
// will be inserted once Execute is called.
func (c *Config) Add() *BotResUpdate {
	return &BotResUpdate{
		db:  c.db,
		ctx: context.Background(),
		log: c.log,
		st:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (c *Config) AddAllDevices(b []BotResInfo) (int64, error) {
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
func (u *BotResUpdate) Execute() error {
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
func (u *BotResUpdate) Serial(serial string) *BotResUpdate {
	u.bresp.SerialNumber = serial

	return u
}

// SlackID will update the value of the users slack id
func (u *BotResUpdate) SlackID(sid string) *BotResUpdate {
	u.bresp.SlackID = sid

	return u
}

// FirstACK will update if the user acknowledged the first message
func (u *BotResUpdate) FirstACK(f bool) *BotResUpdate {
	u.bresp.FirstACK = f

	return u
}

// FirstMessageSent will update if the user was sent the first message
func (u *BotResUpdate) FirstMessageSent(s bool) *BotResUpdate {
	u.bresp.FirstMessageSent = s

	return u
}

// FirstMessageSentAt will update the time the first message was sent
func (u *BotResUpdate) FirstMessageSentAt(t time.Time) *BotResUpdate {
	u.bresp.FirstMessageSentAt = t

	return u
}

// FirstMessageWaiting will update if the user is waiting for the first message
func (u *BotResUpdate) FirstMessageWaiting(w bool) *BotResUpdate {
	u.bresp.FirstMessageWaiting = w

	return u
}

// FirstACKTime will update the time the user acknowledged the first message
func (u *BotResUpdate) FirstACKTime(t time.Time) *BotResUpdate {
	u.bresp.FirstACKTime = t

	return u
}

// FullName will update the users full name
func (u *BotResUpdate) FullName(n string) *BotResUpdate {
	u.bresp.FullName = n

	return u
}

// DelayAt will update the time the user delayed
func (u *BotResUpdate) DelayAt(t time.Time) *BotResUpdate {
	u.bresp.DelayAt = t

	return u
}

// DelayDate will update the date for reminder
func (u *BotResUpdate) DelayDate(d string) *BotResUpdate {
	u.bresp.DelayDate = d

	return u
}

// DelayTime will update the time for reminder
func (u *BotResUpdate) DelayTime(d string) *BotResUpdate {
	u.bresp.DelayTime = d

	return u
}

// DelaySent will update the status for the delay. If sent it will be true.
func (u *BotResUpdate) DelaySent(d bool) *BotResUpdate {
	u.bresp.DelaySent = d

	return u
}

// ManagerSlackID will update the value for the managers slack id
func (u *BotResUpdate) ManagerID(m string) *BotResUpdate {
	u.bresp.ManagerSlackID = m

	return u
}

// ManagerMessageSent will update the value for the manager message sent
func (u *BotResUpdate) ManagerMessageSent(m bool) *BotResUpdate {
	u.bresp.ManagerMessageSent = m

	return u
}

// ManagerMessageSentAt will update the value for the manager message sent at
func (u *BotResUpdate) ManagerMessageSentAt(m time.Time) *BotResUpdate {
	u.bresp.ManagerMessageSentAt = m

	return u
}

// UserEmail will update the value for the users email
func (u *BotResUpdate) UserEmail(e string) *BotResUpdate {
	u.bresp.UserEmail = e

	return u
}
