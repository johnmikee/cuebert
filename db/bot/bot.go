package bot

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// BotResInfo  represents the columns in the devices table
type BotResInfo struct {
	SlackID              string    `json:"slack_id"`
	UserEmail            string    `json:"user_email"`
	ManagerSlackID       string    `json:"manager_slack_id"`
	FirstACK             bool      `json:"first_ack"`
	FirstACKTime         time.Time `json:"first_ack_time"`
	FirstMessageSent     bool      `json:"first_message_sent"`
	FirstMessageSentAt   time.Time `json:"first_message_sent_at"`
	FirstMessageWaiting  bool      `json:"first_message_waiting"`
	ManagerMessageSent   bool      `json:"manager_message_sent"`
	ManagerMessageSentAt time.Time `json:"manager_message_sent_at"`
	FullName             string    `json:"full_name"`
	DelayAt              time.Time `json:"delay_at"`
	DelayDate            string    `json:"delay_date"`
	DelayTime            string    `json:"delay_time"`
	DelaySent            bool      `json:"delay_sent"`
	SerialNumber         string    `json:"serial_number"`
	TZOffset             int64     `json:"tz_offset"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type BR []BotResInfo

func (b BR) Empty() bool {
	return len(b) == 0
}

type Config struct {
	db  *pgxpool.Conn
	ctx context.Context
	log logger.Logger
	st  sq.StatementBuilderType
}

const table = "bot_results"

var columns = []string{
	"slack_id",
	"user_email",
	"manager_slack_id",
	"first_ack",
	"first_ack_time",
	"first_message_sent",
	"first_message_sent_at",
	"first_message_waiting",
	"manager_message_sent",
	"manager_message_sent_at",
	"full_name",
	"delay_at",
	"delay_date",
	"delay_time",
	"delay_sent",
	"serial_number",
	"tz_offset",
	"created_at",
	"updated_at",
}

// Device returns a new client used to interact with the devices table
func Bot(d *pgxpool.Pool, l *logger.Logger) *Config {
	conn, err := d.Acquire(context.Background())
	if err != nil {
		l.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Config{
		db:  conn,
		ctx: context.Background(),
		log: logger.ChildLogger("db/bot", l),
		st:  sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}
