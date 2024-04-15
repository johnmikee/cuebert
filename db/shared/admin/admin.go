package admin

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Info represents a row in the admin table
type Info struct {
	Week           int       `json:"week"`
	ActiveCampaign bool      `json:"active_campaign"`
	ActiveTask     string    `json:"active_task"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UI is a convenience type for a slice of Info
type UI []Info

// Empty checks if the Info slice is empty
func (u UI) Empty() bool {
	return len(u) == 0
}

type Config struct {
	db    *pgxpool.Conn
	ctx   context.Context
	log   logger.Logger
	st    sq.StatementBuilderType
	table string
}

var columns = []string{
	"week",
	"active_campaign",
	"active_task",
	"updated_at",
}

// Column is a convenience type for the column names
type Column string

const (
	Week           Column = "week"
	Task           Column = "task"
	ActiveCampaign Column = "active_campaign"
	ActiveTask     Column = "active_task"
	UpdatedAt      Column = "updated_at"
)

// Admin returns a new client used to interact with the admin table
func Admin(c *config.Config) *Config {
	conn, err := c.DB.Pool.Acquire(context.Background())
	if err != nil {
		c.Log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Config{
		db:    conn,
		ctx:   context.Background(),
		log:   logger.ChildLogger("db/admin", c.Log),
		st:    sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		table: c.Table,
	}
}
