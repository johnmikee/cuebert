package device

import (
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/db/devices"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Device is the internal struct for the device package
type Device struct {
	db     *db.DB
	client mdm.Provider
	log    logger.Logger
}

// Config is the configuration for the Device struct.
type Config struct {
	Client mdm.Provider
	DB     *db.DB
	Log    *logger.Logger
}

// This is an interface for adding devices based on the mdm provider
type DeviceAdder interface {
	mdm.Provider
	AddAllDevices() error
}

func New(d *Config) *Device {
	return &Device{
		db:     d.DB,
		log:    logger.ChildLogger("internal/device", d.Log),
		client: d.Client,
	}
}

// AddAllDevices will add all devices from the MDM to the DB
func (c *Device) AddAllDevices() error {
	db := devices.Device(c.db, &c.log)

	res, err := c.client.ListDevices()
	if err != nil {
		return err
	}

	machines := devices.DI{}

	for i := range res {
		if res[i].Platform == "Mac" {
			di := devices.Info{
				DeviceID:     res[i].DeviceID,
				DeviceName:   res[i].DeviceName,
				Model:        res[i].Model,
				SerialNumber: res[i].SerialNumber,
				Platform:     res[i].Platform,
				OSVersion:    res[i].OSVersion,
				LastCheckIn:  res[i].LastCheckIn,
				User:         res[i].User.Email,
				UserMDMID:    res[i].User.ID,
			}

			machines = append(machines, di)
		}
	}

	_, err = db.AddAllDevices(machines)

	if err != nil {
		return err
	}

	return nil
}
