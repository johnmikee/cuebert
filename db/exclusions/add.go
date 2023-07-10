package exclusions

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/pkg/errors"
)

type Update struct {
	args  []interface{}
	e     Info
	db    *pgxpool.Conn
	st    sq.StatementBuilderType
	query string
	ctx   context.Context
	log   logger.Logger
}

// Addexclusioninitializes a new UserUpdate struct.
//
// the functions below that are methods of Update
// are used to modify specific fields of the statement that
// will be inserted once Execute is called.
func (c *Config) Add() *Update {
	return &Update{
		db:  c.db,
		ctx: context.Background(),
		log: c.log,
		st:  c.st,
	}
}

// Execute sends the statement to add the user after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (e *Update) Execute() (*pgxpool.Conn, error) {
	query, args, err := e.st.Insert(table).
		Columns(columns...).
		Values(
			e.e.Approved,
			e.e.SerialNumber,
			e.e.UserEmail,
			e.e.Reason,
			e.e.Until,
			helpers.UpdateTime(),
			helpers.UpdateTime(),
		).ToSql()

	e.log.Trace().Str("query", query).Interface("args", args).Msg("composed sql")

	if err != nil {
		return nil, errors.Wrap(err, "failed to build insert statement")
	}

	_, err = e.db.Exec(e.ctx, query, args...)

	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	e.db.Release()
	e.log.Info().Msg("input was successfully submitted")
	return e.db, nil
}

// Approved will update the status of the approval
func (e *Update) Approved(a bool) *Update {
	e.e.Approved = a

	return e
}

// Email will update the value of the users email
func (e *Update) Email(email string) *Update {
	e.e.UserEmail = email

	return e
}

// Reason will update the value of the reason for exclusion
func (e *Update) Reason(r string) *Update {
	e.e.Reason = r

	return e
}

// SerialNumber will update the value of the excluded devices serial number
func (e *Update) SerialNumber(s string) *Update {
	e.e.SerialNumber = s

	return e
}

// Until will update the value of the date the exclusion should last until
func (e *Update) Until(t time.Time) *Update {
	e.e.Until = t

	return e
}
