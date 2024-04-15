package pg

import (
	"context"
	"os"
	"runtime"
	"time"

	zerologadapter "github.com/jackc/pgx-zerolog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/pkg/errors"
)

// Connect establishes the connection with the DB.
func (c *Client) Connect() error {
	dbConfig := c.dbConfig()
	if dbConfig == nil {
		return errors.New("db config is nil")
	}

	connPool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return err
	}

	c.pool = connPool
	if err = c.pool.Ping(c.ctx); err != nil {
		c.log.Info().AnErr("error", err).Msg("pinging the db")
		os.Exit(1)
	}
	c.log.Info().Msg("connected to the db")
	return nil
}

// Close implements db.Client.
func (c *Client) Close() error {
	c.pool.Close()

	return nil
}

// GetConnStr implements db.Client.
func (c *Client) GetConnStr() string {
	return c.connStr
}

func (c *Client) dbConfig() *pgxpool.Config {
	const defaultMaxConnLifetime = time.Hour
	const defaultMaxConnIdleTime = time.Minute * 30
	const defaultHealthCheckPeriod = time.Minute
	const defaultConnectTimeout = time.Second * 5

	conn, err := pgxpool.ParseConfig(c.connStr)
	if err != nil {
		return nil
	}

	conn.ConnConfig.Tracer = &tracelog.TraceLog{
		Logger: zerologadapter.NewLogger(logger.New(
			&logger.Config{
				ToFile:  false,
				Level:   logger.TRACE,
				Service: "db-tracer",
				Env:     "prod",
			},
		).Logger),
		LogLevel: tracelog.LogLevel(6),
	}

	conn.MaxConns = int32(runtime.NumCPU() * 5)
	conn.MaxConnLifetime = defaultMaxConnLifetime
	conn.MaxConnIdleTime = defaultMaxConnIdleTime
	conn.HealthCheckPeriod = defaultHealthCheckPeriod
	conn.ConnConfig.ConnectTimeout = defaultConnectTimeout

	return conn
}

func (c *Client) acquire() *pgxpool.Conn {
	conn, err := c.pool.Acquire(context.Background())
	if err != nil {
		c.log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return conn
}
