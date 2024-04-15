package exclusions

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Remove struct {
	db    *pgxpool.Conn
	dt    sq.StatementBuilderType
	sql   sq.DeleteBuilder
	ctx   context.Context
	log   logger.Logger
	table string
}

// Remove initializes a new Remove struct.
//
// the methods of Remove are used to designate specific
// fields of the statement that will be inserted once Execute is called.
func (c *Config) Remove() *Remove {
	return &Remove{
		db:    c.db,
		ctx:   context.Background(),
		log:   c.log,
		dt:    c.st,
		table: c.table,
	}
}

// Remove sends the statement to remove the exclusion after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *Remove) Execute() (*pgxpool.Conn, error) {
	sql, args, err := u.sql.ToSql()

	u.log.Trace().Str("query", sql).Interface("args", args).Msg("composed sql query")
	if err != nil {
		return nil, fmt.Errorf("sql generation failed %w", err)
	}

	_, err = u.db.Exec(
		u.ctx,
		sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exclusion query failed %w", err)
	}

	u.db.Release()
	u.log.Info().Msg("user was successfully removed")

	return u.db, nil
}

func (u *Remove) Where(column Column, val any) *Remove {
	u.sql = u.dt.Delete(u.table).Where(sq.Eq{string(column): val})

	return u
}
