package db

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Table represents the database table.
type Table string

const ( // Available options for the Table type
	BotResults Table = "bot_results"
	Devices    Table = "devices"
)

// Connection stores the database connection
type Connection struct {
	Pool    *pgxpool.Pool
	DB      *pgxpool.Conn
	Log     *logger.Logger
	Table   Table
	Columns []string
	ST      sq.StatementBuilderType
}

// New creates a new database connection.
func New(c *Connection) *Connection {
	conn, err := c.Pool.Acquire(context.Background())
	if err != nil {
		c.Log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Connection{
		DB:    conn,
		Log:   c.Log,
		Table: c.Table,
		ST:    c.ST,
	}
}
