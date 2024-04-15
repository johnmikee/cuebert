package db

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/db/pg"
	_ "github.com/lib/pq"
)

// Client is an interface for database clients.
type Client interface {
	Setup(*config.Config) error
	Connect() error
	Close() error
	GetConnection() *pgxpool.Conn
	GetPool() *pgxpool.Pool
	GetConnStr() string
	Ping() error
}

// New creates a new database client.
func New(c *config.Config) (Client, error) {
	var client Client
	switch c.Type {
	// case "mysql":
	// 	client = &mysql.Client{}
	case "postgres":
		client = &pg.Client{}
	default:
		return nil, errors.New("unsupported database type")
	}
	err := client.Setup(c)

	return client, err
}
