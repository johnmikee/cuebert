package main

import (
	"flag"
	"fmt"
)

func (cmd *DC) removeDevices(args []string) error {
	var err error
	flagset := flag.NewFlagSet("remove", flag.ExitOnError)
	var (
		id       = flagset.String("id", "", "remove a devices based on the device id")
		model    = flagset.String("model", "", "remove devices based on the model")
		name     = flagset.String("name", "", "remove devices based on the hostname")
		platform = flagset.String("platform", "", "remove devices based on the platform")
		os       = flagset.String("os", "", "remove devices with a particular OS version")
		serial   = flagset.String("serial", "", "remove devices based on serial number")
		user     = flagset.String("user", "", "remove devices based on the device user")
		userid   = flagset.String("user-id", "", "remove devices based on the device users id")
	)
	flagset.Usage = usageFor(flagset, "cue devices get db devices [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	base := cmd.db.RemoveDeviceBy()

	count := 0

	if *id != "" {
		base.ID(*id)
		count++
	}
	if *model != "" {
		base.Model(*model)
		count++
	}
	if *name != "" {
		base.Name(*name)
		count++
	}
	if *os != "" {
		base.OS(*os)
		count++
	}
	if *platform != "" {
		base.Platform(*platform)
		count++
	}
	if *serial != "" {
		base.Serial(*serial)
		count++
	}
	if *user != "" {
		base.User(*user)
		count++
	}
	if *userid != "" {
		base.UserMDMID(*id)
		count++
	}
	if count > 1 {
		return fmt.Errorf("only one arg can be set for device removal")
	}

	_, err = base.Execute()

	if err != nil {
		return err
	}

	return nil
}
