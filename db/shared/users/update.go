package users

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/parser"
)

type Update struct {
	args   []interface{}
	ui     Info
	query  string
	config *Config
}

func (c *Config) Update() *Update {
	return &Update{
		config: c,
	}
}

func (u *Update) WithOptions(i Info) *Update {
	u.ui = i

	return u
}

// Update sends the statement to update the device after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *Update) Send() (*pgxpool.Conn, error) {
	_, err := u.config.db.Exec(
		u.config.ctx,
		u.query, u.args...)
	if err != nil {
		return nil, err
	}

	u.config.db.Release()
	u.config.log.Info().Msg("input was successfully submitted")

	return u.config.db, nil
}

func (u *Update) Parse(column Column, val any) *Update {
	check := parser.ExtractCheckInfo(u.ui)
	query, args, err := parser.ParseInput(
		&parser.Parser{
			Index:  string(column),
			Table:  u.config.table,
			Val:    val,
			Check:  check,
			Method: parser.Update,
			Into:   u.ui,
		},
	)
	if err != nil {
		u.config.log.Err(err).Msg("parsing input")
		return nil
	}

	u.args = args
	u.query = query

	return u
}
