package updates

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Info struct {
	Week      int       `json:"week"`
	ItemName  string    `json:"item_name"`
	Monday    int       `json:"monday"`
	Tuesday   int       `json:"tuesday"`
	Wednesday int       `json:"wednesday"`
	Thursday  int       `json:"thursday"`
	Friday    int       `json:"friday"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UI []Info

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

const (
	table        = "updates"
	DefaultTable = table
)

var columns = []string{
	"week",
	"item_name",
	"monday",
	"tuesday",
	"wednesday",
	"thursday",
	"friday",
	"updated_at",
}

func Updates(c *config.Config) *Config {
	conn, err := c.DB.Pool.Acquire(context.Background())
	if err != nil {
		c.Log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Config{
		db:    conn,
		ctx:   context.Background(),
		log:   logger.ChildLogger("db/updates", c.Log),
		st:    c.ST,
		table: c.Table,
	}
}
