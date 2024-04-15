package task

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
	"user_task_id",
	"task_id",
	"user_id",
	"user_email",
	"item_name",
	"task_description",
	"task_complete",
	"updated_at",
}

// Column represents the database column.
type Column string

const (
	UserTaskID      Column = "user_task_id"
	TaskID          Column = "task_id"
	UserID          Column = "user_id"
	UserEmail       Column = "user_email"
	ItemName        Column = "item_name"
	TaskDescription Column = "task_description"
	TaskComplete    Column = "task_complete"
	UpdatedAt       Column = "updated_at"
)

// Info contains the information for a task.
type Info struct {
	UserTaskID      int       `json:"user_task_id"`
	TaskID          int       `json:"task_id"`
	UserID          string    `json:"user_id"`
	UserEmail       string    `json:"user_email"`
	ItemName        string    `json:"item_name"`
	TaskDescription string    `json:"task_description"`
	TaskComplete    bool      `json:"task_complete"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TI is a custom type for a slice of Info for use with helper functions.
type TI []Info

// Empty checks if the response is empty.
func (t TI) Empty() bool {
	return len(t) == 0
}

// UserInt returns the user id as an integer.
func (i Info) UserInt() int {
	uid, err := strconv.Atoi(i.UserID)
	if err != nil {
		return 0
	}

	return uid
}

// Completed returns a slice of completed tasks.
func (t TI) Completed() TI {
	var completed TI
	for i := range t {
		if t[i].TaskComplete {
			completed = append(completed, t[i])
		}
	}
	return completed
}

// Config stores the database connection
type Config struct {
	db    *pgxpool.Conn
	ctx   context.Context
	log   *logger.Logger
	st    sq.StatementBuilderType
	table string
}

// Task returns a new Config struct to interact with the task table.
func Task(c *config.Config) *Config {
	conn, err := c.DB.Acquire(context.Background())
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
