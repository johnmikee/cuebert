package tracking

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/parser"
)

// Update is a struct to hold config for updating an item in the tracking table.
type Update struct {
	args   []interface{}
	ui     Info
	query  string
	config *Config
}

// Update returns a new Update struct.
func (c *Config) Update() *Update {
	return &Update{
		config: c,
	}
}

// WithOptions adds the options to query.
func (u *Update) WithOptions(i Info) *Update {
	u.ui = i

	return u
}

// Send takes the prepared statement and executes it.
func (u *Update) Send() (*pgxpool.Conn, error) {
	_, err := u.config.db.Exec(
		u.config.ctx,
		u.query, u.args...)
	if err != nil {
		return nil, err
	}

	u.config.db.Release()

	return u.config.db, nil
}

// Parse takes the input and parses it into a query.
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
		u.config.log.Err(err).Msg("parse input")
		return nil
	}

	u.args = args
	u.query = query

	return u
}
