package config

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Config is used to configure the connection to the DB.
type Config struct {
	Connection Connection
	DB         DB
	Type       Implementation // ("mysql", "postgres", "sqlite")
	TaskName   string
	Table      string
	Log        *logger.Logger
	ST         squirrel.StatementBuilderType
}

// Connection stores the connection information for the DB.
type Connection struct {
	Host     string
	Name     string
	Password string
	Port     string
	User     string
	SSLMode  string
	TimeZone string
}

// DB stores the connection pool and connection to the DB.
type DB struct {
	DB   *pgxpool.Conn
	Pool *pgxpool.Pool
}

// Acquire will return a connection from the pool.
func (d *DB) Acquire(ctx context.Context) (*pgxpool.Conn, error) {
	if d.Pool != nil {
		return d.Pool.Acquire(ctx)
	}

	return nil, errors.New("pool is nil")
}

// Implementation is a convenience type for the supported DB types.
type Implementation string

const (
	MySQL    Implementation = "mysql"
	Postgres Implementation = "postgres"
	SQLite   Implementation = "sqlite"
)

// Validate will check the required arguments to establish a connection.
func (c *Config) Validate() (bool, error) {
	if c.Connection.User == "" {
		return false, errors.New("must include user value")
	}

	if c.Connection.Password == "" {
		return false, errors.New("must include password value")
	}

	if c.Connection.Host == "" {
		return false, errors.New("must include host value")
	}

	if c.Connection.Name == "" {
		return false, errors.New("must include dbname value")
	}

	if c.Connection.Port == "" {
		return false, errors.New("must include port value")
	}

	return true, nil
}

// UpdateOptions is a function type to update the Config struct.
type UpdateOptions func(*Config) *Config

// Update will update the Config struct with the provided options.
func (c *Config) Update(opts ...UpdateOptions) *Config {
	for _, opt := range opts {
		c = opt(c)
	}
	return c
}

// WithTable is an option to update the Config struct with a table name.
func WithTable(table string) UpdateOptions {
	return func(c *Config) *Config {
		c.Table = table
		return c
	}
}
