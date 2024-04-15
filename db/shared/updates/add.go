package updates

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
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
		ai:     Info{},
	}
}

func (u *Add) BatchAdd(us []Info) (int64, error) {
	rows := [][]interface{}{}
	for _, u := range us {
		row := []interface{}{
			u.Week,
			u.ItemName,
			u.Monday,
			u.Tuesday,
			u.Wednesday,
			u.Thursday,
			u.Friday,
			helpers.UpdateTime(),
		}
		rows = append(rows, row)
	}

	copyCount, err := u.config.db.CopyFrom(
		context.Background(),
		pgx.Identifier{u.config.table},
		columns,
		pgx.CopyFromRows(rows))

	u.config.db.Release()

	return copyCount, err
}

func (u *Add) Execute() error {
	query, args, err := u.config.st.Insert(u.config.table).
		Columns(columns...).
		Values(
			u.ai.Week,
			u.ai.ItemName,
			u.ai.Monday,
			u.ai.Tuesday,
			u.ai.Wednesday,
			u.ai.Thursday,
			u.ai.Friday,
			helpers.UpdateTime(),
		).
		ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build sql statement")
	}

	u.config.log.Trace().Str("query", query).Interface("args", args).Msg("composed sql query")

	_, err = u.config.db.Exec(
		u.config.ctx,
		query,
		args...)
	if err != nil {
		return err
	}

	u.config.log.Info().Msg("input was successfully submitted")
	u.config.log.Trace().
		Str("table", u.config.table).
		Interface("response", u.ai).
		Msg("inserted into table")

	u.config.db.Release()

	return nil
}

func (u *Add) SetDay(c int) *Add {
	today := int(time.Now().Weekday())
	switch today {
	case 1:
		u.ai.Monday = c
	case 2:
		u.ai.Tuesday = c
	case 3:
		u.ai.Wednesday = c
	case 4:
		u.ai.Thursday = c
	case 5:
		u.ai.Friday = c
	}
	return u
}

func (u *Add) WithOptions(i Info) *Add {
	u.ai = i

	return u
}
