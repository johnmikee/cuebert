package task

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// Add contains the information for adding a new task.
type Add struct {
	ui     Info
	config *Config
}

// Add returns a new Add struct.
func (c *Config) Add() *Add {
	return &Add{
		config: c,
		ui:     Info{},
	}
}

// All adds all the tasks to the database.
func (u *Add) All(t TI) (int64, error) {
	rows := [][]interface{}{}
	for i := range t {
		row := []interface{}{
			t[i].UserID,
			t[i].TaskID,
			t[i].UserEmail,
			t[i].ItemName,
			t[i].TaskDescription,
			t[i].TaskComplete,
			t[i].UpdatedAt,
		}
		rows = append(rows, row)
	}
	copyCount, err := u.config.db.CopyFrom(
		context.Background(),
		pgx.Identifier{u.config.table},
		[]string{"user_id", "task_id", "user_email", "item_name", "task_description", "task_complete", "updated_at"},
		pgx.CopyFromRows(rows))

	u.config.db.Release()

	return copyCount, err
}

// Execute takes the prepared statement and executes it.
func (u *Add) Execute() error {
	query, args, err := u.config.st.Insert(u.config.table).
		Columns("user_id", "task_id", "user_email", "item_name", "task_description", "task_complete", "updated_at").
		Values(
			u.ui.UserID,
			u.ui.TaskID,
			u.ui.UserEmail,
			u.ui.ItemName,
			u.ui.TaskDescription,
			u.ui.TaskComplete,
			u.ui.UpdatedAt,
		).
		ToSql()
	if err != nil {
		return err
	}

	_, err = u.config.db.Exec(
		u.config.ctx,
		query, args...)
	if err != nil {
		return err
	}

	u.config.db.Release()

	return nil
}

// WithOptions sets the options for the Add struct.
func (a *Add) WithOptions(i Info) *Add {
	a.ui = i

	return a
}
