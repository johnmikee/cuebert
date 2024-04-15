package devices

import (
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

func (u *Update) Parse(column string, val any) *Update {
	u.config.log.Trace().Str("query", u.query).Interface("args", u.args).Msg("composed sql query")

	check := parser.ExtractCheckInfo(u.ui)
	query, args, err := parser.ParseInput(
		&parser.Parser{
			Index:  column,
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
