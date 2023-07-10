package users

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Info represents the columns in the users table
type Info struct {
	MDMID        string     `json:"user_mdm_id"`
	UserLongName string     `json:"user_long_name"`
	UserEmail    string     `json:"user_email"`
	UserSlackID  string     `json:"user_slack_id"`
	TZOffset     int64      `json:"tz_offset"`
	CreatedAt    *time.Time `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type UI []Info

func (u UI) Empty() bool {
	return len(u) == 0
}

// Config is used to interact with the users table
type Config struct {
	db  *pgxpool.Conn
	ctx context.Context
	log logger.Logger
	st  sq.StatementBuilderType
}

const table = "users"

var columns = []string{
	"user_mdm_id",
	"user_long_name",
	"user_email",
	"user_slack_id",
	"tz_offset",
	"created_at",
	"updated_at",
}

// User returns a new client used to interact with the users table
func User(d *db.DB, l *logger.Logger) *Config {
	conn, err := d.Acquire(context.Background())
	if err != nil {
		l.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Config{
		db:  conn,
		ctx: context.Background(),
		log: logger.ChildLogger("db/users", l),
		st:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}
