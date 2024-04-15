package exclusions

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Info represents the columns in the exclusions table
type Info struct {
	Approved     bool      `json:"approved"`
	SerialNumber string    `json:"serial_number"`
	UserEmail    string    `json:"user_email"`
	Reason       string    `json:"reason"`
	UpdatedAt    time.Time `json:"updated_at"`
	Until        time.Time `json:"until"`
}

type EI []Info

func (e EI) Empty() bool {
	return len(e) == 0
}

type Config struct {
	columns []string
	db      *pgxpool.Conn
	ctx     context.Context
	log     logger.Logger
	st      sq.StatementBuilderType
	table   string
}

var osColumns = []string{
	"approved",
	"serial_number",
	"user_email",
	"reason",
	"until",
	"updated_at",
}

var complianceColumns = []string{
	"approved",
	"user_email",
	"reason",
	"until",
	"updated_at",
}

type Column string

const (
	Approved     Column = "approved"
	SerialNumber Column = "serial_number"
	UserEmail    Column = "user_email"
	Reason       Column = "reason"
	Until        Column = "until"
)

// Exclusion returns a new client used to interact with the exclusions table
func Exclusion(c *config.Config) *Config {
	conn, err := c.DB.Pool.Acquire(context.Background())
	if err != nil {
		c.Log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	var columns []string
	switch c.TaskName {
	case string(implementation.OS):
		columns = osColumns
	default:
		columns = complianceColumns
	}

	return &Config{
		columns: columns,
		ctx:     context.Background(),
		db:      conn,
		log:     logger.ChildLogger(c.Table, c.Log),
		st:      sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		table:   c.Table,
	}
}
