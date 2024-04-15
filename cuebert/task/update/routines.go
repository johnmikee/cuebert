package update

import (
	"time"

	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/cuebert/task/info"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/cuebert/task/update/common"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tables"
)

// ClientTasks is the interface for the update client.
type ClientTasks interface {
	task.Messaging
	CheckComplianceStatus(user string) (bool, error)
	CheckComplianceAll() ([]info.PersistUsers, error) // this should return a slice of info.Users who should be present in the database
	GetTasks() ([]info.Tasks, error)
	GetUsers() ([]info.Users, error)
	Setup(*task.Config, *common.Common, *tables.Config)
}

// Get implements task.Rabbit.
func (c *Client) Get() []task.Job {
	return []task.Job{
		{
			Job:  c.Check,
			Name: "Check",
			When: time.Duration(c.config.Flags.CheckInterval) * time.Minute,
		},
		{
			Job:  c.Diff,
			Name: "Diff",
			When: time.Duration(5) * time.Minute,
		},
		{
			Job:  c.Poll,
			Name: "Poll",
			When: time.Duration(c.config.Flags.PollInterval) * time.Minute,
		},
	}
}

// Check implements task.Rabbit.
func (c *Client) Check(t time.Time) {
	c.common.Check(t,
		&shared.ReminderPayload{
			FirstMessage: c.impl.FirstMessage(),
			FinalMessage: c.impl.FinalMessage(),
			Text:         c.impl.FirstMessage(),
			Title:        c.impl.FirstMessageHeader(),
		},
	)
}

// CheckCompliance implements task.Rabbit.
func (c *Client) CheckCompliance(time.Time) {
	c.shared.CheckCompliance(time.Now())
}

// Diff implements task.Rabbit.
func (c *Client) Diff(t time.Time) {
	c.common.Diff(t)
}

// Poll implements task.Rabbit.
func (c *Client) Poll(t time.Time) {
	c.common.Poll(t)
}

// Populate implements task.Rabbit.
func (c *Client) Populate(time.Time) error {
	users, err := c.impl.GetUsers()
	if err != nil {
		return err
	}

	c.log.Info().Interface("users", users).Msg("users")
	tasks, err := c.impl.GetTasks()
	if err != nil {
		return err
	}
	c.log.Info().Interface("tasks", tasks).Msg("tasks")
	_, err = c.common.TableBuilder(
		&common.TableBuilder{
			Users: users,
			Tasks: tasks,
		},
	)

	return err
}
