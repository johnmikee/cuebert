package admin

import (
	"github.com/johnmikee/cuebert/db/parser"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

type Update struct {
	args   []interface{}
	query  string
	config *Config
	ui     Info
}

func (c *Config) Update() *Update {
	return &Update{
		config: c,
	}
}

func (u *Update) Send() error {
	query, args, err := u.config.st.Insert(u.config.table).
		Columns(columns...).
		Values(
			u.ui.Week,
			u.ui.ActiveCampaign,
			u.ui.ActiveTask,
			helpers.UpdateTime(),
		).ToSql()
	if err != nil {
		return err
	}

	_, err = u.config.db.Exec(u.config.ctx, query, args...)
	if err != nil {
		return err
	}

	u.config.db.Release()
	u.config.log.Info().Msg("input was successfully submitted")
	return nil
}

func (u *Update) WithOptions(i Info) *Update {
	u.ui = i

	return u
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
		u.config.log.Err(err).Msg("parse input")
		return nil
	}

	u.args = args
	u.query = query

	return u
}
