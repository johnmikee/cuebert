package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/johnmikee/cuebert/db/devices"
)

func (cmd *DC) addusage() {
	const help = `cue devices add:

Valid Options:
  * all
  * some

Use cue devices add <option> -h for additional usage of each command.
Example: cue devices add some -h
	`
	fmt.Print(help)
}

func (cmd *DC) addDevices(args []string) error {
	if len(args) < 1 {
		cmd.addusage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "all":
		err := cmd.devices.AddAllDevices()
		cmd.run = func(s []string) error { return err }

	case "some":
		cmd.run = cmd.addSome

	default:
		cmd.addusage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

// addSome allows passing either a single item or comma separated list for each option.
//
// the devices passed for the flag will be queried against the mdm and then inserted
// into the db.
//
// alternatively, each field can be passed as a flag. this requires passing all required
// fields which will be validated.for certain flags (id, serial, and user) we can fall back
// on pulling from the MDM if all fields are not present. for the remaining flags if all the
// flags are not present we will return an error.
func (cmd *DC) addSome(args []string) error {
	var err error
	flagset := flag.NewFlagSet("add", flag.ExitOnError)
	var (
		id       = flagset.String("id", "", "device ids (comma-separated)")
		model    = flagset.String("model", "", "device models (comma-separated)")
		name     = flagset.String("name", "", "device hostnames (comma-separated)")
		platform = flagset.String("platform", "", "device platform (comma-separated)")
		os       = flagset.String("os", "", "device OS version (comma-separated)")
		serial   = flagset.String("serial", "", "device serial numbers (comma-separated)")
		user     = flagset.String("user", "", "device user (comma-separated)")
	)
	flagset.Usage = usageFor(flagset, "cue devices add some [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	required := []string{*id, *model, *name, *platform, *os, *serial, *user}

	info := &devices.DeviceInfo{
		DeviceID:     *id,
		DeviceName:   *name,
		Model:        *model,
		SerialNumber: *serial,
		Platform:     *platform,
		OSVersion:    *os,
		User:         *user,
	}

	if *id != "" {
		if !validateAddArgs(required) {
			dev, err := cmd.mdm.GetDevice(*id)
			if err != nil {
				return fmt.Errorf("could not find device in kandji %s", err)
			}
			ki := &devices.DeviceInfo{
				DeviceID:     dev.DeviceID,
				DeviceName:   dev.DeviceName,
				Model:        dev.Model,
				SerialNumber: dev.SerialNumber,
				Platform:     dev.Platform,
				LastCheckIn:  dev.LastCheckIn,
				OSVersion:    dev.OSVersion,
				User:         dev.User.Name,
			}
			err = cmd.addDevice(ki)

			return err
		}
	}

	if *serial != "" || *user != "" {
		dev, err := cmd.mdm.ListDevices()
		if err != nil {
			return err
		}
		for i := range dev {
			if *user == dev[i].User.Email || *serial == dev[i].SerialNumber {
				di := &devices.DeviceInfo{
					DeviceID:     dev[i].DeviceID,
					DeviceName:   dev[i].DeviceName,
					Model:        dev[i].Model,
					SerialNumber: dev[i].SerialNumber,
					Platform:     dev[i].Platform,
					LastCheckIn:  dev[i].LastCheckIn,
					OSVersion:    dev[i].OSVersion,
					User:         dev[i].User.Email,
				}
				err = cmd.addDevice(di)

				return err
			}
		}
	}
	if *model != "" || *name != "" || *os != "" || *platform != "" {
		if !validateAddArgs(required) {
			return err
		}
		err = cmd.addDevice(info)
	}

	return err
}

func (cmd *DC) addDevice(d *devices.DeviceInfo) error {
	_, err := cmd.db.AddDevice().
		ID(d.DeviceID).
		Model(d.Model).
		OS(d.OSVersion).
		Platform(d.Platform).
		Serial(d.SerialNumber).
		User(d.User).
		Execute()

	return err
}
