package devices

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// DeviceInfo represents the columns in the devices table
type DeviceInfo struct {
	CreatedAt    *time.Time `json:"created_at"`
	DeviceID     string     `json:"device_id"`
	DeviceName   string     `json:"device_name"`
	LastCheckIn  *time.Time `json:"last_check_in"`
	Model        string     `json:"model"`
	OSVersion    string     `json:"os_version"`
	Platform     string     `json:"platform"`
	SerialNumber string     `json:"serial_number"`
	User         string     `json:"user"`
	UserMDMID    string     `john:"user_mdm_id"`
	UpdatedAt    *time.Time `json:"updated_at"`
}

type DI []DeviceInfo

func (d DI) Empty() bool {
	return len(d) == 0
}

type Config struct {
	ctx context.Context
	db  *pgxpool.Conn
	log logger.Logger
	st  sq.StatementBuilderType
}

const table = "devices"

var columns = []string{
	"device_id",
	"device_name",
	"model",
	"serial_number",
	"platform",
	"os_version",
	"user_name",
	"user_mdm_id",
	"last_check_in",
	"created_at",
	"updated_at",
}

// Device returns a new client used to interact with the devices table
func Device(d *pgxpool.Pool, l *logger.Logger) *Config {
	conn, err := d.Acquire(context.Background())
	if err != nil {
		l.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Config{
		db:  conn,
		ctx: context.Background(),
		log: logger.ChildLogger("db/device", l),
		st:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}
