package update

import (
	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/cuebert/task/update/common"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tables"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/cuebert/task/update/os"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Client implements task.Rabbit.
type Client struct {
	common *common.Common
	config *task.Config
	impl   ClientTasks
	idp    idp.Provider
	log    logger.Logger
	shared *shared.Shared
	tables *tables.Config
}

// CheckNotSent makes sure we arent duplicating messages for users with multiple devices.
func (c *Client) CheckNotSent(serviceID string, slackID string) (bool, error) {
	ti, err := c.tables.Tracking().By().Opts(tracking.UserSlackID, slackID).Query()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not get tracking info")
		return true, err
	}
	if ti.Empty() {
		c.log.Debug().Msg("no tracking info found")
		return true, nil
	}

	for _, t := range ti {
		c.log.Trace().Interface("tracking", t).Msg("checking tracking")
		if t.FinalMessageSent { // return false if we have already sent the final message
			return false, nil
		}
	}

	c.log.Trace().Msg("no final message sent")
	return true, nil
}

// FinalMessageSent implements task.Rabbit.
func (c *Client) FinalMessageSent(serviceID string, slackID string) error {
	return c.common.FinalMessageSent(serviceID, slackID)
}

// IsExcluded implements task.Rabbit.
func (c *Client) IsExcluded(serial string) (bool, bool) {
	return c.tables.IsExcluded(serial)
}

// Setup implements task.Rabbit.
func (c *Client) Setup(t *task.Config) {
	share := shared.New(&shared.Config{
		Config:      t.Config,
		Bot:         t.Bot,
		LifeCycle:   t.LifeCycle,
		DB:          t.DB,
		Flags:       t.Flags,
		Log:         t.Log,
		Status:      t.StatusHandler,
		Caller:      c,
		ImplChecker: c.impl,
	})

	c.shared = share
	tables := tables.New(
		&tables.Tables{
			DB:     t.DB,
			Config: t.DBConfig,
			Log:    t.Log,
		},
	)
	c.tables = tables
	common := common.New(
		&common.Config{
			Bot:    t.Bot,
			Config: t.Config,
			Flags:  t.Flags,
			Log:    t.Log,
			Tables: tables,
			Status: t.StatusHandler,
			Shared: share,
		},
	)
	c.common = common
	c.config = t
	c.log = logger.ChildLogger("os-update", &t.Log)
	impl := new(os.Client)
	impl.Setup(t, common, tables)
	c.impl = impl

	c.shared.ImplChecker(c.impl)

	c.log.Info().Msg("shared update setup complete")
}
