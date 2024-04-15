package updates

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/parser"
)

type Update struct {
	config *Config
	query  string
	args   []interface{}
	ui     Info
	table  string
}

func (c *Config) Update() *Update {
	return &Update{
		config: c,
		table:  c.table,
	}
}

func (u *Update) Send() (*pgxpool.Conn, error) {
	u.config.log.Trace().Str("query", u.query).Interface("args", u.args).Msg("composed sql query")
	_, err := u.config.db.Exec(
		u.config.ctx,
		u.query, u.args...)
	if err != nil {
		return nil, err
	}

	u.config.db.Release()
	u.config.log.Trace().Str("table", u.config.table).Msg("input was successfully submitted")

	return u.config.db, nil
}

func (u *Update) WithOptions(i Info) *Update {
	u.ui = i

	return u
}

func (u *Update) Parse(column string, val any) *Update {
	check := parser.ExtractCheckInfo(u.ui)
	query, args, err := parser.ParseInput(
		&parser.Parser{
			Index:  column,
			Table:  u.table,
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
