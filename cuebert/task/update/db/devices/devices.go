package devices

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Info represents the columns in the devices table.
type Info struct {
	DeviceID     string     `json:"device_id"`
	DeviceName   string     `json:"device_name"`
	LastCheckIn  *time.Time `json:"last_check_in"`
	Model        string     `json:"model"`
	OSVersion    string     `json:"os_version"`
	Platform     string     `json:"platform"`
	SerialNumber string     `json:"serial_number"`
	User         string     `json:"device_user"`
	UserMDMID    string     `john:"user_mdm_id"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type DI []Info

func (d DI) Empty() bool {
	return len(d) == 0
}

type Config struct {
	ctx   context.Context
	db    *pgxpool.Conn
	log   logger.Logger
	st    sq.StatementBuilderType
	table string
}

var columns = []string{
	"device_id",
	"device_name",
	"model",
	"serial_number",
	"platform",
	"os_version",
	"device_user",
	"user_mdm_id",
	"last_check_in",
	"updated_at",
}

type Column string

const (
	DeviceID     Column = "device_id"
	DeviceName   Column = "device_name"
	LastCheckIn  Column = "last_check_in"
	Model        Column = "model"
	OSVersion    Column = "os_version"
	Platform     Column = "platform"
	SerialNumber Column = "serial_number"
	User         Column = "user"
	UserMDMID    Column = "user_mdm_id"
	UpdatedAt    Column = "updated_at"
)

// Device returns a new client used to interact with the devices table.
func Device(c *config.Config) *Config {
	conn, err := c.DB.Pool.Acquire(context.Background())
	if err != nil {
		c.Log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Config{
		db:    conn,
		ctx:   context.Background(),
		log:   logger.ChildLogger("db/device", c.Log),
		st:    c.ST,
		table: c.Table,
	}
}
