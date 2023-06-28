package users

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type UserRemove struct {
	sql sq.DeleteBuilder
	st  sq.StatementBuilderType
	db  *pgxpool.Conn

	ctx context.Context
	log logger.Logger
}

// RemoveUser initializes a new UserRemove struct.
//
// the methods of UserRemove are used to designate specific
// fields of the statement that will be inserted once Execute is called.
func (c *Config) Remove() *UserRemove {
	c.log.Info().Msg("removing user")
	return &UserRemove{
		db:  c.db,
		ctx: context.Background(),
		log: c.log,
		st:  c.st,
	}
}

// Run sends the statement to remove the user after it has been composed.
//
// returns the connection which should be closed after checking the error.
func (u *UserRemove) Run() (*pgxpool.Conn, error) {
	sql, args, err := u.sql.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = u.db.Exec(
		u.ctx,
		sql, args...)

	if err != nil {
		return nil, err
	}

	u.db.Release()
	u.log.Info().Msg("user was successfully removed")

	return u.db, nil
}

// All will remove all rows from the table
func (u *UserRemove) All() *UserRemove {
	u.sql = u.st.Delete("").From(table)
	return u
}

// ID will remove the row based off the user_mdm_id for the user
func (u *UserRemove) ID(id ...string) *UserRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_mdm_id": id})
	return u
}

// Email remove the row based off the user_email for the user
func (u *UserRemove) Email(email ...string) *UserRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_email": email})
	return u
}

// UserLongName will remove the row based off the user_long_name for the user
func (u *UserRemove) UserLongName(name ...string) *UserRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_long_name": name})
	return u
}

// SlackID will remove the row based off the slack_id for the user
func (u *UserRemove) SlackID(s ...string) *UserRemove {
	u.sql = u.st.Delete(table).Where(sq.Eq{"user_slack_id": s})
	return u
}
