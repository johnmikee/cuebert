package common

import (
	"time"

	"github.com/johnmikee/cuebert/db/shared/updates"
)

// CS is a struct to hold the task name, status, and count.
type CS struct {
	Task   string
	Status bool
	Count  int
}

// CSCount is a custom type for a slice of CS for use with helper functions.
type CSCount []CS

// CompleteCount returns a slice of completed tasks.
func (c *Common) CompleteCount() CSCount {
	all, err := c.tables.Task().By().All().Query()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not query devices")
		return nil
	}

	var count CSCount
	for _, task := range all {
		count = append(count, CS{
			Task:   task.ItemName,
			Status: task.TaskComplete,
		})
	}

	return count
}

// AddUpdates pulls the completed tasks for use in the daily updates.
func (s CSCount) AddUpdates(c *Common) error {
	_, week := time.Now().ISOWeek()
	day := int(time.Now().Weekday())

	complete := 0

	for _, m := range s {
		if m.Status {
			complete++
		}
	}

	u := updates.Info{}
	if len(s) > 0 {
		u.Week = week
		u.ItemName = s[0].Task
		switch day {
		case 1:
			u.Monday = complete
		case 2:
			u.Tuesday = complete
		case 3:
			u.Wednesday = complete
		case 4:
			u.Thursday = complete
		case 5:
			u.Friday = complete
		}
	}

	ui, err := c.tables.Updates().By().Day(updates.Today).Week(week).Build().Query()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not query updates")
		return err
	}

	if ui.Empty() {
		err = s.add(u, c)
	} else {
		err = s.update(u, c)
	}

	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not add updates")
		return err
	}

	return nil
}

func (s CSCount) add(u updates.Info, c *Common) error {
	return c.tables.Updates().Add().WithOptions(u).Execute()
}

func (s CSCount) update(u updates.Info, c *Common) error {
	_, err := c.tables.Updates().
		Update().WithOptions(u).
		Parse("week", u.Week).Send()

	return err
}
