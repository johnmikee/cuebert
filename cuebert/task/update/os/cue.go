package os

import (
	"strconv"

	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"
	"github.com/johnmikee/cuebert/cuebert/task/info"
	"github.com/johnmikee/cuebert/mdm"

	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/task/update/common"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/devices"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tables"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Client struct {
	bot    *bot.Bot
	mdm    mdm.Provider
	flags  *config.Flags
	tables *tables.Config
	log    logger.Logger
}

func (c *Client) Setup(t *task.Config, cc *common.Common, tc *tables.Config) {
	c.bot = t.Bot
	c.mdm = t.MDM
	c.flags = t.Flags
	c.tables = tc
	c.log = logger.ChildLogger("os", &t.Log)
}

// CheckComplianceAll implements compliance.Compliance.
func (c *Client) CheckComplianceAll() ([]info.PersistUsers, error) {
	devices, err := c.mdm.ListDevices()
	if err != nil {
		return nil, err
	}

	users := []info.PersistUsers{}
	for _, d := range devices {
		ok, _ := helpers.CompareOSVer(d.OSVersion, c.flags.RequiredVers)
		if !ok {
			c.log.Debug().
				Str("os", d.OSVersion).
				Str("serial", d.SerialNumber).
				Str("device", d.DeviceName).
				Msg("device is out of compliance")
			users = append(users, info.PersistUsers{
				UserID: func() int {
					uid, err := strconv.Atoi(d.User.ID)
					if err != nil {
						return 0
					}
					return uid
				}(),
				UserEmail: d.User.Email,
				Task:      implementation.OS,
			})
		}
	}

	return users, nil
}

// CheckComplianceStatus implements update.Messaging.
func (c *Client) CheckComplianceStatus(user string) (bool, error) {
	ti, err := c.tables.Tracking().By().Opts(tracking.UserSlackID, user).Query()
	if err != nil {
		return false, err
	}
	if ti.Empty() {
		return false, nil
	}

	deviceID, err := c.tables.Device().By().Opts(devices.SerialNumber, ti[0].SerialNumber).Query()
	if err != nil {
		return false, err
	}
	if deviceID.Empty() {
		return false, nil
	}

	device, err := c.mdm.GetDevice(deviceID[0].DeviceID)
	if err != nil {
		return false, err
	}

	return helpers.CompareOSVer(device.OSVersion, c.flags.RequiredVers)
}

func (c *Client) GetTasks() ([]info.Tasks, error) {
	c.log.Info().Msg("getting tasks...")
	return nil, nil
}

func (c *Client) GetUsers() ([]info.Users, error) {
	c.log.Info().Msg("getting users...")
	return nil, nil
}
