package tables

import (
	"github.com/johnmikee/cuebert/cuebert/task/update/db/devices"
	"github.com/johnmikee/cuebert/db/shared/users"
)

// DevicesByUser returns all devices for a user from the devices table.
func (c *Config) ByUser(user string) (devices.DI, error) {
	userInfo, err := c.User().By().Opts(users.UserTaskID, user).Query()
	if err != nil {
		return nil, err
	}

	users := []int{}
	for _, i := range userInfo {
		users = append(users, i.UserServiceID)
	}

	s, err := c.Device().By().Opts(devices.UserMDMID, users[0]).Query()
	if err != nil {
		return s, err
	}

	return s, nil
}
