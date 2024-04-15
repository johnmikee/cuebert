package exclusions

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/parser"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Update struct {
	args  []interface{}
	e     Info
	db    *pgxpool.Conn
	st    sq.StatementBuilderType
	query string
	ctx   context.Context
	log   logger.Logger
	table string
}

// Update initializes a new Update struct.
//
// The methods of Update are used to modify which values
// will be updated.
//
// Those are parsed and the column we are using as the condition to match is passed as the index.
func (c *Config) Update() *Update {
	return &Update{
		db:    c.db,
		ctx:   context.Background(),
		log:   c.log,
		st:    c.st,
		table: c.table,
	}
}

// Update sends the statement to update the device after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *Update) Send() (*pgxpool.Conn, error) {
	_, err := u.db.Exec(
		u.ctx,
		u.query, u.args...)
	if err != nil {
		return nil, err
	}

	u.db.Release()
	u.log.Info().Msg("input was successfully submitted")
	return u.db, nil
}

func (u *Update) WithOptions(i Info) *Update {
	u.e = i

	return u
}

func (u *Update) Parse(column string, val any) *Update {
	check := parser.ExtractCheckInfo(u.e)
	query, args, err := parser.ParseInput(
		&parser.Parser{
			Index:  column,
			Table:  u.table,
			Val:    val,
			Check:  check,
			Method: parser.Update,
			Into:   u.e,
		},
	)
	if err != nil {
		u.log.Err(err).Msg("parse input")
		return nil
	}

	u.args = args
	u.query = query

	return u
}
