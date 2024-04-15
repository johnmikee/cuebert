package users

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Info represents the columns in the users table
type Info struct {
	UserServiceID int       `json:"user_service_id"`
	UserLongName  string    `json:"user_long_name"`
	UserEmail     string    `json:"user_email"`
	UserSlackID   string    `json:"user_slack_id"`
	TZOffset      int64     `json:"tz_offset"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UI []Info

func (u UI) Empty() bool {
	return len(u) == 0
}

func (u UI) GetServiceID() int {
	return u[0].UserServiceID
}

// Config is used to interact with the users table
type Config struct {
	db    *pgxpool.Conn
	ctx   context.Context
	log   logger.Logger
	st    sq.StatementBuilderType
	table string
}

const table = "users"

var DefaultTable = table

var columns = []string{
	"user_service_id",
	"user_long_name",
	"user_email",
	"user_slack_id",
	"tz_offset",
	"updated_at",
}

type Column string

const (
	UserTaskID   Column = "user_service_id"
	UserLongName Column = "user_long_name"
	UserEmail    Column = "user_email"
	UserSlackID  Column = "user_slack_id"
	TZOffset     Column = "tz_offset"
	UpdatedAt    Column = "updated_at"
)

// User returns a new client used to interact with the users table
func User(c *config.Config) *Config {
	c.Log.Info().Msg("connecting to db...")
	conn, err := c.DB.Pool.Acquire(context.Background())
	if err != nil {
		c.Log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	c.Log.Info().Msg("connected to db")

	return &Config{
		db:    conn,
		ctx:   context.Background(),
		log:   logger.ChildLogger("db/users", c.Log),
		st:    c.ST,
		table: c.Table,
	}
}
