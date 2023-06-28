package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/johnmikee/cuebert/db/devices"
	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/internal/device"
	"github.com/johnmikee/cuebert/internal/user"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type DC struct {
	devices *device.Device
	users   *user.User
	log     logger.Logger
	mdm     mdm.Provider
	c       *Config
	db      *db.Config
	run     func([]string) error
}

func (out *TableOutput) deviceHeader() {
	fmt.Fprintf(out.w, "device_id\tdevice_name\tmodel\tserial_number\tplatform\tos_version\tuser_name\tlast_check_in\tcreated_at\tupdated_at\n")
}

func (out *TableOutput) deviceMDMHeader() {
	fmt.Fprintf(out.w, "asset_tag\tdevice_id\tdevice_name\tmodel\tos_version\tserial_number\tplatform\tuser_email\tuser_id\tuser_name\n")
}

func (cmd *DC) usage() {
	const help = `cue devices:

Valid Options:
  * add
  * get
  * remove
  * update

Use cue devices <option> -h for additional usage of each command.
Example: cue devices get -h
	`
	fmt.Print(help)
}

func (c *CueConfig) Devices(args []string) error {
	cmd := &DC{
		c:       c.conf,
		db:      c.db,
		log:     c.log,
		devices: c.devices,
		users:   c.users,
	}

	if len(args) < 1 {
		cmd.usage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "add":
		cmd.run = cmd.addDevices

	case "get":
		cmd.run = cmd.getDeviceSource

	case "remove":
		cmd.run = cmd.removeDevices

	case "update":
		cmd.run = cmd.updateDevices

	default:
		cmd.usage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func devicePrinter(devices []devices.DeviceInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 9, 2, ' ', 0)
	out := &TableOutput{w}
	out.deviceHeader()
	defer out.basicFooter()
	for i := range devices {
		fmt.Fprintf(out.w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			devices[i].DeviceID,
			devices[i].DeviceName,
			devices[i].Model,
			devices[i].SerialNumber,
			devices[i].Platform,
			devices[i].OSVersion,
			devices[i].User,
			devices[i].LastCheckIn,
			devices[i].CreatedAt,
			devices[i].UpdatedAt)
	}
}

func deviceMDMPrinter(devices mdm.DeviceResults) {
	w := tabwriter.NewWriter(os.Stdout, 0, 9, 2, ' ', 0)
	out := &TableOutput{w}
	out.deviceMDMHeader()
	defer out.basicFooter()
	for i := range devices {
		fmt.Fprintf(out.w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			devices[i].AssetTag,
			devices[i].DeviceID,
			devices[i].DeviceName,
			devices[i].Model,
			devices[i].OSVersion,
			devices[i].SerialNumber,
			devices[i].Platform,
			devices[i].User.Email,
			devices[i].User.ID,
			devices[i].User.Name)
	}
}
