package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/johnmikee/cuebert/db/devices"
	"github.com/johnmikee/cuebert/mdm"
)

func (cmd *DC) sourceusage() {
	const help = `cue devices get:

Valid Options:
  * db
  * mdm

Use cue devices get [option] -h for additional usage of each command.
Example: cue devices get db -h
	`
	fmt.Print(help)
}

func (cmd *DC) getDeviceSource(args []string) error {
	if len(args) < 1 {
		cmd.sourceusage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "db":
		cmd.run = cmd.getDevicesDB
	case "mdm":
		cmd.run = cmd.getDevicesMDM

	default:
		cmd.sourceusage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func (cmd *DC) getDevicesDB(args []string) error {
	flagset := flag.NewFlagSet("db", flag.ExitOnError)
	var (
		all      = flagset.Bool("all", false, "return all devices")
		id       = flagset.String("id", "", "return a devices based on the device id (comma-separated)")
		model    = flagset.String("model", "", "return devices based on the model (comma-separated)")
		name     = flagset.String("name", "", "return devices based on the hostname (comma-separated)")
		platform = flagset.String("platform", "", "return devices based on the platform (comma-separated)")
		os       = flagset.String("os", "", "return devices with a particular OS version (comma-separated)")
		serial   = flagset.String("serial", "", "return devices based on serial number (comma-separated)")
		user     = flagset.String("user", "", "return devices based on the device user (comma-separated)")
		userid   = flagset.String("user-id", "", "return devices based on the device users id (comma-separated)")
	)
	flagset.Usage = usageFor(flagset, "cue devices get db [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	var di devices.DI
	var err error

	if *all {
		_, err = cmd.db.GetAllDevices()
	}

	base := cmd.db.QueryDeviceBy()

	if *id != "" {
		idSlice := strings.Split(*id, ",")
		di, err = base.ID(idSlice...).Query()
	}

	if *model != "" {
		modelSlice := strings.Split(*model, ",")
		di, err = base.Model(modelSlice...).Query()
	}

	if *name != "" {
		nameSlice := strings.Split(*name, ",")
		di, err = base.Name(nameSlice...).Query()
	}

	if *os != "" {
		osSlice := strings.Split(*os, ",")
		di, err = base.OS(osSlice...).Query()
	}

	if *platform != "" {
		platformSlice := strings.Split(*os, ",")
		di, err = base.Platform(platformSlice...).Query()
	}

	if *serial != "" {
		serialSlice := strings.Split(*serial, ",")
		di, err = base.Serial(serialSlice...).Query()
	}

	if *user != "" {
		userSlice := strings.Split(*user, ",")
		di, err = base.User(userSlice...).Query()
	}

	if *userid != "" {
		useridSlice := strings.Split(*userid, ",")
		di, err = base.UserID(useridSlice...).Query()
	}

	if err != nil {

		return err
	}

	devicePrinter(di)
	return nil
}

func (cmd *DC) getDevicesMDM(args []string) error {
	flagset := flag.NewFlagSet("mdm", flag.ExitOnError)
	var (
		all      = flagset.Bool("all", false, "return all devices")
		assetTag = flagset.String("assetTag", "", "return a devices based on the asset tag (comma-separated)")
		deviceID = flagset.String("deviceID", "", "return devices based on the device id (comma-separated)")
		name     = flagset.String("deviceName", "", "return devices based on the hostname (comma-separated)")
		model    = flagset.String("model", "", "return devices based on the model (comma-separated)")
		platform = flagset.String("platform", "", "return devices based on the platform (comma-separated)")
		os       = flagset.String("os", "", "return devices with a particular OS version (comma-separated)")
		serial   = flagset.String("serial", "", "return devices based on serial number (comma-separated)")
		user     = flagset.String("user", "", "return devices based on the device user (comma-separated)")
		userid   = flagset.String("user-id", "", "return devices based on the device users id (comma-separated)")
		username = flagset.String("username", "", "return devices based on the device users name (comma-separated)")
	)
	flagset.Usage = usageFor(flagset, "cue devices get mdm [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}
	var mr mdm.DeviceResults
	var err error

	if *all {
		mr, err = cmd.mdm.ListDevices()
		if err != nil {
			return err
		}
		deviceMDMPrinter(mr)
	}

	// if any of the other flags are not empty fill the query opts and send it

	if *assetTag != "" || *deviceID != "" || *name != "" || *platform != "" || *os != "" || *serial != "" || *user != "" || *userid != "" || *username != "" {
		mr, err = cmd.mdm.QueryDevices(
			&mdm.QueryOpts{
				AssetTag:     *assetTag,
				DeviceID:     *deviceID,
				DeviceName:   *name,
				Model:        *model,
				OSVersion:    *os,
				SerialNumber: *serial,
				Platform:     *platform,
				UserEmail:    *user,
				UserID:       *userid,
				UserName:     *username,
			},
		)
		if err != nil {
			return err
		}

		deviceMDMPrinter(mr)
	}

	return nil
}
