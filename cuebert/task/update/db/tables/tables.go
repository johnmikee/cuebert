package tables

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/devices"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/db/shared/admin"
	"github.com/johnmikee/cuebert/db/shared/exclusions"
	"github.com/johnmikee/cuebert/db/shared/updates"
	"github.com/johnmikee/cuebert/db/shared/users"

	"github.com/johnmikee/cuebert/pkg/logger"
)

type Check struct {
	Serial string
	OS     string
}

// Config is a struct to hold config for connecting to the
// different tables in the DB.
type Config struct {
	admin      func(*config.Config) *admin.Config
	devices    func(*config.Config) *devices.Config
	exclusions func(*config.Config) *exclusions.Config
	updates    func(*config.Config) *updates.Config
	user       func(*config.Config) *users.Config
	tracking   func(*config.Config) *tracking.Config

	log        logger.Logger
	connection *config.Config
}

func a(s *config.Config) *admin.Config {
	s.Table = s.TaskName + "_os_admin"
	return admin.Admin(s)
}

// Admin returns the admin config for connecting to the admin table in the DB.
func (c *Config) Admin() *admin.Config {
	return c.admin(c.connection)
}

func d(c *config.Config) *devices.Config {
	c.Table = c.TaskName + "_os_devices"
	return devices.Device(c)
}

func (c *Config) Device() *devices.Config {
	return c.devices(c.connection)
}

func e(c *config.Config) *exclusions.Config {
	c.Table = c.TaskName + "_os_exclusions"
	return exclusions.Exclusion(c)
}

// Exclusion returns the exclusion config for connecting to the exclusions table in the DB.
func (c *Config) Exclusion() *exclusions.Config {
	return c.exclusions(c.connection)
}

func t(s *config.Config) *tracking.Config {
	s.Table = s.TaskName + "_os_tracking"
	return tracking.Tracking(s)
}

// Tracking returns the tracking config for connecting to the tracking table in the DB.
func (c *Config) Tracking() *tracking.Config {
	return c.tracking(c.connection)
}

func u(s *config.Config) *users.Config {
	s.Table = s.TaskName + "_os_users"
	return users.User(s)
}

// User returns the user config for connecting to the users table in the DB.
func (c *Config) User() *users.Config {
	return c.user(c.connection)
}

func up(s *config.Config) *updates.Config {
	s.Update(config.WithTable(s.TaskName + "_os_updates"))
	return updates.Updates(s)
}

// Updates returns the updates config for connecting to the updates table in the DB.
func (c *Config) Updates() *updates.Config {
	return c.updates(c.connection)
}

// Tables is a struct to hold config for connecting to the different tables in the DB.
type Tables struct {
	DB     db.Client
	Config *config.Config
	Log    logger.Logger
}

// New returns a new db connection
func New(s *Tables) *Config {
	return &Config{
		admin:      a,
		devices:    d,
		exclusions: e,
		updates:    up,
		user:       u,
		tracking:   t,
		log:        logger.ChildLogger("tables", &s.Log),
		connection: s.Config,
	}
}

func (c *Config) GetConnection() *pgxpool.Conn {
	return c.connection.DB.DB
}

func (c *Config) Pool() *pgxpool.Pool {
	return c.connection.DB.Pool
}

// returns a bool indicating if the serial number is in the exclusions table and a bool for the approval status
func (c *Config) IsExcluded(serial string) (bool, bool) {
	ex, err := c.Exclusion().By().Opts(exclusions.SerialNumber, serial).Query()
	if err != nil {
		c.log.Err(err).Msg("could not check if device was in exclusions table")
		return false, false
	}

	if ex.Empty() {
		return false, false
	}

	if ex[0].SerialNumber == serial {
		return true, ex[0].Approved
	}

	return false, false
}
