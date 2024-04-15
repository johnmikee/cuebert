package update

import (
	"time"

	"github.com/johnmikee/cuebert/db/shared/admin"
)

func (c *Client) Start() {
	_, week := time.Now().ISOWeek()
	_, err := c.tables.Admin().Add().
		WithOptions(admin.Info{
			Week:           week,
			ActiveCampaign: true,
			ActiveTask:     c.config.Flags.TaskImplementation,
		}).Execute()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("failed to update admin table")
	}
}

func (c *Client) Stop() {
	c.config.LifeCycle.Stop()
}

func (c *Client) Update() {
	c.config.LifeCycle.Update()
}
