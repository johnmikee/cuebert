package users

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
	args  []interface{}
	user  Info
	query string
	db    *pgxpool.Conn
	st    sq.StatementBuilderType
	ctx   context.Context
	log   logger.Logger
}

// AddAllUsers should only be used in specific cases.
//
// - to initialize the DB with values
// - if a table is dropped and rebuilt
//
// This will add all users passed to the DB.
func (c *Config) AddAllUsers(us []Info) (int64, error) {
	rows := [][]interface{}{}
	for _, u := range us {
		if u.MDMID != "" {
			row := []interface{}{
				u.MDMID,
				u.UserLongName,
				u.UserEmail,
				u.UserSlackID,
				u.TZOffset,
				helpers.UpdateTime(),
				helpers.UpdateTime(),
			}
			rows = append(rows, row)
		}
	}

	copyCount, err := c.db.CopyFrom(
		context.Background(),
		pgx.Identifier{table},
		columns,
		pgx.CopyFromRows(rows))

	c.db.Release()

	return copyCount, err

}

// Add initializes a new Update struct.
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

// Execute sends the statement to add the user after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *Update) Execute() (*pgxpool.Conn, error) {
	query, args, err := u.st.Insert(table).
		Columns(columns...).
		Values(
			u.user.MDMID,
			u.user.UserLongName,
			u.user.UserEmail,
			u.user.UserSlackID,
			u.user.TZOffset,
			helpers.UpdateTime(),
			helpers.UpdateTime()).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	_, err = u.db.Exec(
		u.ctx,
		query,
		args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	u.db.Release()
	u.log.Info().Msg("input was successfully submitted")
	return u.db, nil
}

// Email will update the value of the users email
func (u *Update) Email(user string) *Update {
	u.user.UserEmail = user

	return u
}

// ID will update the value of the users id
func (u *Update) ID(id string) *Update {
	u.user.MDMID = id

	return u
}

// LongName will update the value of the users full name
func (u *Update) LongName(ln string) *Update {
	u.user.UserLongName = ln

	return u
}

// Slack will update the value of the users slack id
func (u *Update) Slack(user string) *Update {
	u.user.UserSlackID = user

	return u
}

// TZ will update the value of the users tz_offset
func (u *Update) TZ(tz int64) *Update {
	u.user.TZOffset = tz

	return u
}
