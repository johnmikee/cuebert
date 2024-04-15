package common

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/johnmikee/cuebert/cuebert/task/info"
	"github.com/johnmikee/cuebert/db/shared/admin"
)

// TableBuilder is a struct to hold the data needed to build the tables.
type TableBuilder struct {
	Users []info.Users
	Tasks []info.Tasks
}

// TableBuilder implements task.Rabbit.
func (c *Common) TableBuilder(t *TableBuilder) ([]string, error) {
	c.log.Info().Msg("initializing tables...")

	check, err := c.AddAllUsers(t.Users)
	if err != nil {
		c.log.Err(err).Msg("could not add user table")
	}
	err = c.AddAllTasks(t.Tasks)
	if err != nil {
		c.log.Err(err).Msg("could not add tasks table")
	}

	// we need this built before we can do anything else
	err = c.BuildTrackingTable()

	return check, err
}

func userTask(taskimpl string) (string, []interface{}, error) {
	return sq.StatementBuilder.PlaceholderFormat(sq.Dollar).
		Select("cu.user_email, cu.user_long_name, cu.user_slack_id, cu.tz_offset, ct.task_complete").
		From(fmt.Sprintf("%s_compliance_task ct", taskimpl)).
		Join(fmt.Sprintf("%s_compliance_users cu ON (cu.user_email = ct.user_email)", taskimpl)).
		ToSql()
}

// Overlap stores the information when querying the overlap between tables.
type Overlap struct {
	UserEmail    string `json:"user_email"`
	UserLongName string `json:"user_long_name"`
	UserSlackID  string `json:"user_slack_id"`
	TZOffset     int64  `json:"tz_offset"`
	TaskComplete bool   `json:"task_complete"`
}

func (c *Common) overlap() ([]Overlap, error) {
	query, args, err := userTask(c.flags.TaskImplementation)
	if err != nil {
		return nil, err
	}
	c.log.Trace().Str("query", query).Interface("args", args).Msg("composed sql query")

	rows, err := c.tables.GetConnection().Query(context.Background(), query, args...)
	if err != nil {
		c.log.Trace().AnErr("error", err).Msg("querying db")
		return nil, err
	}

	add := []Overlap{}

	for rows.Next() {
		var overlap Overlap

		err = rows.Scan(
			&overlap.UserEmail,
			&overlap.UserLongName,
			&overlap.UserSlackID,
			&overlap.TZOffset,
			&overlap.TaskComplete,
		)
		if err != nil {
			return nil, fmt.Errorf("row query failed %w", err)
		}

		add = append(add, overlap)

	}

	c.log.Info().Interface("overlap", add).Msg("overlap")

	return add, nil
}

// BuildTrackingTable builds the tracking table.
func (c *Common) BuildTrackingTable() error {
	overlap, err := c.overlap()
	if err != nil {
		return err
	}
	c.log.Info().Interface("overlap", overlap).Msg("overlap")
	var updates []tracking.Info

	for i := range overlap {
		if overlap[i].TaskComplete {
			c.log.Trace().Str("user", overlap[i].UserEmail).Msg("user has already completed task")
			continue
		}
		updates = append(updates, tracking.Info{
			TaskComplete: overlap[i].TaskComplete,
			UserSlackID:  overlap[i].UserSlackID,
			UserEmail:    overlap[i].UserEmail,
			UserFullName: overlap[i].UserLongName,
			TZOffset:     overlap[i].TZOffset,
		})

		if c.flags.TaskImplementation == "ethena" {
			_, err = c.tables.Tracking().Update().WithOptions(tracking.Info{
				TaskComplete: overlap[i].TaskComplete,
				UserSlackID:  overlap[i].UserSlackID,
				UserEmail:    overlap[i].UserEmail,
				UserFullName: overlap[i].UserLongName,
				TZOffset:     overlap[i].TZOffset,
			}).Parse(tracking.UserEmail, overlap[i].UserEmail).Send()
			if err != nil {
				c.log.Debug().AnErr("error", err).Msg("updating tracking table")
				return err
			}
		}
	}
	if c.flags.TaskImplementation != "ethena" {
		_, err = c.tables.Tracking().Add().All(updates)
		if err != nil {
			c.log.Debug().AnErr("error", err).Msg("adding users to tracking table")
			return err
		}
	}

	_, week := time.Now().ISOWeek()
	_, err = c.tables.Admin().Add().WithOptions(
		admin.Info{
			Week:           week,
			ActiveCampaign: true,
			ActiveTask:     c.flags.TaskImplementation,
			UpdatedAt:      time.Time{},
		}).Execute()

	return err
}
