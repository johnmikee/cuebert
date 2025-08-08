package tables

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/task"
	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/db/shared/admin"
	"github.com/johnmikee/cuebert/db/shared/exclusions"
	"github.com/johnmikee/cuebert/db/shared/updates"
	"github.com/johnmikee/cuebert/db/shared/users"

	"github.com/johnmikee/cuebert/pkg/logger"
)

// Config is a struct to hold config for connecting to the different tables in the DB.
type Config struct {
	admin      func(*config.Config) *admin.Config
	exclusions func(*config.Config) *exclusions.Config
	tracking   func(*config.Config) *tracking.Config
	task       func(*config.Config) *task.Config
	updates    func(*config.Config) *updates.Config
	user       func(*config.Config) *users.Config

	log        logger.Logger
	client     db.Client
	connection *config.Config
}

func u(s *config.Config) *users.Config {
	s.Table = s.TaskName + "_compliance_users"
	return users.User(s)
}

// User returns the user config for connecting to the users table in the DB.
func (c *Config) User() *users.Config {
	return c.user(c.connection)
}

func a(s *config.Config) *admin.Config {
	s.Update(config.WithTable("compliance_admin"))
	return admin.Admin(s)
}

// Admin returns the admin config for connecting to the admin table in the DB.
func (c *Config) Admin() *admin.Config {
	return c.admin(c.connection)
}

func e(c *config.Config) *exclusions.Config {
	c.Table = c.TaskName + "_compliance_exclusions"
	return exclusions.Exclusion(c)
}

// Exclusion returns the exclusion config for connecting to the exclusions table in the DB.
func (c *Config) Exclusion() *exclusions.Config {
	return c.exclusions(c.connection)
}

func r(s *config.Config) *tracking.Config {
	s.Table = s.TaskName + "_compliance_tracking"
	return tracking.Tracking(s)
}

// Tracking returns the tracking config for connecting to the tracking table in the DB.
func (c *Config) Tracking() *tracking.Config {
	return c.tracking(c.connection)
}

func t(s *config.Config) *task.Config {
	s.Table = s.TaskName + "_compliance_task"
	return task.Task(s)
}

// Task returns the task config for connecting to the task table in the DB.
func (c *Config) Task() *task.Config {
	return c.task(c.connection)
}

func up(s *config.Config) *updates.Config {
	s.Update(config.WithTable(s.TaskName + "_compliance_updates"))
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

// New returns a new db connection.
func New(s *Tables) *Config {
	return &Config{
		user:       u,
		exclusions: e,
		task:       t,
		admin:      a,
		updates:    up,
		tracking:   r,
		log:        logger.ChildLogger("tables", &s.Log),
		client:     s.DB,
		connection: s.Config,
	}
}

func (c *Config) GetConnection() *pgxpool.Conn {
	return c.connection.DB.DB
}
