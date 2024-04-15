package admin

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/pkg/errors"
)

type Add struct {
	ai     Info
	config *Config
}

func (c *Config) Add() *Add {
	return &Add{
		config: c,
	}
}

// Execute sends the statement to add the user after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (e *Add) Execute() (*pgxpool.Conn, error) {
	query, args, err := e.config.st.Insert(e.config.table).
		Columns(columns...).
		Values(
			e.ai.Week,
			e.ai.ActiveCampaign,
			e.ai.ActiveTask,
			helpers.UpdateTime(),
		).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build insert statement")
	}

	_, err = e.config.db.Exec(e.config.ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}

	e.config.db.Release()
	e.config.log.Info().Msg("input was successfully submitted")
	return e.config.db, nil
}

func (a *Add) WithOptions(i Info) *Add {
	a.ai = i

	return a
}
