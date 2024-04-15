package users

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/pkg/errors"
)

type Add struct {
	ui     Info
	config *Config
}

// AddAll should only be used in specific cases.
//
// - to initialize the DB with values
// - if a table is dropped and rebuilt
//
// This will add all users passed to the DB.
func (c *Add) All(us []Info) (int64, error) {
	for _, u := range us {
		query, args, err := c.config.st.Insert(c.config.table).Columns(columns...).
			Values(
				u.UserServiceID,
				u.UserLongName,
				u.UserEmail,
				u.UserSlackID,
				u.TZOffset,
				helpers.UpdateTime()).Suffix("ON CONFLICT DO NOTHING").ToSql()
		if err != nil {
			c.config.log.Debug().AnErr("error", err).Msg("could not build query")
			continue
		}

		_, err = c.config.db.Exec(context.Background(), query, args...)
		if err != nil {
			c.config.log.Debug().AnErr("error", err).Msg("could not execute query")
			continue
		}
	}

	return 0, nil
}

// Add initializes a new Update struct.
//
// the functions below that are methods of Update
// are used to modify specific fields of the statement that
// will be inserted once Execute is called.
func (c *Config) Add() *Add {
	return &Add{
		config: c,
		ui:     Info{},
	}
}

// Execute sends the statement to add the user after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *Add) Execute() (*pgxpool.Conn, error) {
	query, args, err := u.config.st.Insert(u.config.table).
		Columns(columns...).
		Values(
			u.ui.UserServiceID,
			u.ui.UserLongName,
			u.ui.UserEmail,
			u.ui.UserSlackID,
			u.ui.TZOffset,
			helpers.UpdateTime()).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	_, err = u.config.db.Exec(
		u.config.ctx,
		query,
		args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	u.config.db.Release()
	u.config.log.Info().Msg("input was successfully submitted")
	return u.config.db, nil
}

func (a *Add) WithOptions(i Info) *Add {
	a.ui = i

	return a
}
