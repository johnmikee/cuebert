package tracking

import (
	"context"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/pkg/logger"
)

var columns = []string{
	"user_service_id",
	"task_complete",
	"task_name",
	"user_slack_id",
	"user_email",
	"user_full_name",
	"first_ack",
	"first_ack_time",
	"first_message_sent",
	"first_message_sent_at",
	"first_message_waiting",
	"delay_at",
	"delay_date",
	"delay_time",
	"reminder_sent",
	"reminder_interval",
	"reminder_waiting",
	"tz_offset",
	"updated_at",
	"final_message_sent",
	"final_message_sent_at",
}

// Column represents the tables columns.
type Column string

const (
	UserServiceID       Column = "user_service_id"
	TaskComplete        Column = "task_complete"
	TaskName            Column = "task_name"
	UserSlackID         Column = "user_slack_id"
	UserEmail           Column = "user_email"
	UserFullName        Column = "user_full_name"
	FirstACK            Column = "first_ack"
	FirstACKTime        Column = "first_ack_time"
	FirstMessageSent    Column = "first_message_sent"
	FirstMessageSentAt  Column = "first_message_sent_at"
	FirstMessageWaiting Column = "first_message_waiting"
	DelayAt             Column = "delay_at"
	DelayDate           Column = "delay_date"
	DelayTime           Column = "delay_time"
	ReminderSent        Column = "reminder_sent"
	ReminderInterval    Column = "reminder_interval"
	ReminderWaiting     Column = "reminder_waiting"
	TZOffset            Column = "tz_offset"
	UpdatedAt           Column = "updated_at"
	FinalMessageSent    Column = "final_message_sent"
	FinalMessageSentAt  Column = "final_message_sent_at"
)

// Info contains the information for a task.
type Info struct {
	UserServiceID       string    `json:"user_service_id"`
	TaskComplete        bool      `json:"task_complete"`
	TaskName            string    `json:"task_name"`
	UserSlackID         string    `json:"user_slack_id"`
	UserEmail           string    `json:"user_email"`
	UserFullName        string    `json:"user_full_name"`
	FirstACK            bool      `json:"first_ack"`
	FirstACKTime        time.Time `json:"first_ack_time"`
	FirstMessageSent    bool      `json:"first_message_sent"`
	FirstMessageSentAt  time.Time `json:"first_message_sent_at"`
	FirstMessageWaiting bool      `json:"first_message_waiting"`
	DelayAt             time.Time `json:"delay_at"`
	DelayDate           string    `json:"delay_date"`
	DelayTime           string    `json:"delay_time"`
	ReminderSent        bool      `json:"reminder_sent"`
	ReminderInterval    int       `json:"reminder_interval"`
	ReminderWaiting     bool      `json:"reminder_waiting"`
	TZOffset            int64     `json:"tz_offset"`
	UpdatedAt           time.Time `json:"updated_at"`
	FinalMessageSent    bool      `json:"final_message_sent"`
	FinalMessageSentAt  time.Time `json:"final_message_sent_at"`
}

type TI []Info

func (u TI) Empty() bool {
	return len(u) == 0
}

func (i Info) UserInt() int {
	uid, err := strconv.Atoi(i.UserServiceID)
	if err != nil {
		return 0
	}

	return uid
}

// EmailSlackIDMap returns a map of user emails to slack ids
func (u TI) EmailSlackIDMap() map[string]string {
	m := make(map[string]string)
	for _, v := range u {
		m[v.UserEmail] = v.UserSlackID
	}
	return m
}

type Config struct {
	db    *pgxpool.Conn
	ctx   context.Context
	log   *logger.Logger
	st    sq.StatementBuilderType
	table string
}

// Tracking returns a new tracking config
func Tracking(c *config.Config) *Config {
	conn, err := c.DB.Pool.Acquire(context.Background())
	if err != nil {
		c.Log.Info().AnErr("acquiring connection", err).Msg("failed to acquire lock")
		return nil
	}

	return &Config{
		db:    conn,
		ctx:   context.Background(),
		log:   c.Log,
		st:    c.ST,
		table: c.Table,
	}
}
