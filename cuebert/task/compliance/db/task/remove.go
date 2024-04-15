package task

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Remove contains the information for removing a task.
type Remove struct {
	sql    sq.DeleteBuilder
	config *Config
}

// Remove returns a new Remove struct.
func (c *Config) Remove() *Remove {
	return &Remove{
		config: c,
	}
}

// Execute takes the prepared statement and executes it.
func (u *Remove) Execute() (*pgxpool.Conn, error) {
	sql, args, err := u.sql.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = u.config.db.Exec(
		u.config.ctx,
		sql, args...)
	if err != nil {
		return nil, err
	}

	u.config.db.Release()
	u.config.log.Info().Msg("user was successfully removed")

	return u.config.db, nil
}

// All removes all the tasks from the database.
func (u *Remove) All() *Remove {
	u.sql = u.config.st.Delete("").From(u.config.table)
	return u
}

// Where adds a where clause to the query.
func (u *Remove) Where(column Column, val any) *Remove {
	u.sql = u.config.st.Delete(u.config.table).Where(sq.Eq{string(column): val})

	return u
}
