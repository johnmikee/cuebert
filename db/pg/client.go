package pg

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/pkg/logger"
	_ "github.com/lib/pq"
)

// Client implements db.Client.
type Client struct {
	connStr string
	ctx     context.Context
	pool    *pgxpool.Pool
	db      *pgxpool.Conn
	log     logger.Logger
	st      sq.StatementBuilderType
}

// GetPool implements db.Client.
func (c *Client) GetPool() *pgxpool.Pool {
	return c.pool
}

// GetConnection implements db.Client.
func (c *Client) GetConnection() *pgxpool.Conn {
	return c.db
}

// Ping implements db.Client.
func (c *Client) Ping() error {
	return c.pool.Ping(c.ctx)
}

// Setup implements db.Client.
func (c *Client) Setup(cfg *config.Config) error {
	ok, err := cfg.Validate()
	if !ok {
		return err
	}

	c.ctx = context.Background()
	c.connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.Connection.User, cfg.Connection.Password, cfg.Connection.Host, cfg.Connection.Port, cfg.Connection.Name, cfg.Connection.SSLMode)
	c.log = logger.ChildLogger("postgres", cfg.Log)
	c.st = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	err = c.Connect()
	if err != nil {
		return err
	}
	c.db = c.acquire()

	return nil
}
