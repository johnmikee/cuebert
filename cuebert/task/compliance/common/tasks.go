package common

import (
	"strconv"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/task"
	"github.com/johnmikee/cuebert/cuebert/task/info"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

// AddAllTasks adds all tasks to the DB.
func (c *Common) AddAllTasks(ct []info.Tasks) error {
	ti := []task.Info{}
	for _, t := range ct {
		c.log.Info().Interface("task", t).Msg("adding task")
		ti = append(ti,
			task.Info{
				UserID:          strconv.Itoa(t.UserID),
				ItemName:        t.ItemName,
				TaskDescription: t.TaskDescription,
				TaskComplete:    t.TaskComplete,
				UpdatedAt:       helpers.UpdateTime(),
			},
		)
	}

	var err error
	for _, t := range ti {
		err = c.tables.Task().Add().WithOptions(t).Execute()
		if err != nil {
			c.log.Err(err).Msg("updating tasks")
			continue
		}
	}

	return err
}

// UpdateAllTasks updates all tasks in the DB.
func (c *Common) UpdateAllTasks(ct []info.Tasks) error {
	var err error
	for _, t := range ct {
		c.log.Info().Interface("task", t).Msg("updating task")
		_, err = c.tables.Task().Update().WithOptions(
			task.Info{
				UserID:          strconv.Itoa(t.UserID),
				ItemName:        t.ItemName,
				TaskDescription: t.TaskDescription,
				TaskComplete:    t.TaskComplete,
			}).Parse(task.UserID, strconv.Itoa(t.UserID)).Send()
		if err != nil {
			c.log.Err(err).Msg("updating tasks")
			continue
		}
	}
	return err
}
