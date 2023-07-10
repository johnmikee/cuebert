package tables

import (
	"context"
	"strconv"

	"github.com/johnmikee/cuebert/cuebert/device"
	"github.com/johnmikee/cuebert/cuebert/user"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/db/devices"
	"github.com/johnmikee/cuebert/db/exclusions"
	"github.com/johnmikee/cuebert/db/users"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Config is a struct to hold config for connecting to the
// different tables in the DB.
type Config struct {
	user       func(*db.DB, *logger.Logger) *users.Config
	exclusions func(*db.DB, *logger.Logger) *exclusions.Config
	dev        func(*db.DB, *logger.Logger) *devices.Config
	br         func(*db.DB, *logger.Logger) *bot.Config

	db      *db.DB
	log     logger.Logger
	devices *device.Device
	users   *user.User
}

func b(db *db.DB, l *logger.Logger) *bot.Config {
	return bot.Bot(db, l)
}

func e(db *db.DB, l *logger.Logger) *exclusions.Config {
	return exclusions.Exclusion(db, l)
}

func d(db *db.DB, l *logger.Logger) *devices.Config {
	return devices.Device(db, l)
}

func u(db *db.DB, l *logger.Logger) *users.Config {
	return users.User(db, l)
}

// New returns a new db connection
func New(db *db.DB, log *logger.Logger, opts ...Option) *Config {
	config := &Config{
		user:       u,
		exclusions: e,
		dev:        d,
		br:         b,
		log:        logger.ChildLogger("tables", log),
		db:         db,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

type Option func(*Config)

func WithUsers(u *user.User) func(*Config) {
	return func(c *Config) {
		c.users = u
	}
}

func WithDevices(d *device.Device) func(*Config) {
	return func(c *Config) {
		c.devices = d
	}
}

type Conf = db.DBConfig
type Conn = db.PGDB

func Connect(c *Conf) (*Conn, error) {
	conn, err := db.NewConn(c).Connect()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (db *Config) TestConnection() error {
	return db.db.Ping(context.Background())
}

type Get string

const (
	Host     Get = "host"
	Password Get = "password"
	Name     Get = "name"
	Port     Get = "port"
	UserName Get = "username"
)

// Print out args for connection to the DB
func (db *Config) Print(item Get) string {
	switch item {
	case Host:
		return db.db.Config().ConnConfig.Host
	case Name:
		return db.db.Config().ConnConfig.Database
	case Password:
		return db.db.Config().ConnConfig.Password
	case Port:
		return strconv.Itoa(int(db.db.Config().ConnConfig.Port))
	case UserName:
		return db.db.Config().ConnConfig.User
	default:
		return ""
	}
}
