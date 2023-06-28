package main

import (
	"flag"
	"fmt"
)

func (cmd *DC) updateDevices(args []string) error {
	var err error
	flagset := flag.NewFlagSet("update", flag.ExitOnError)

	var (
		condition    = flagset.String("condition", "", "the condition to base the condition off [required]")
		conditionVal = flagset.String("condition-val", "", "the value of the condition [required]")
		id           = flagset.String("id", "", "device id")
		model        = flagset.String("model", "", "device model")
		name         = flagset.String("name", "", "device hostname")
		platform     = flagset.String("platform", "", "device platform")
		os           = flagset.String("os", "", "device OS version")
		serial       = flagset.String("serial", "", "device serial number")
		user         = flagset.String("user", "", "device user")
		userid       = flagset.String("user-id", "", "device user id")
	)
	flagset.Usage = usageFor(flagset, "cue devices update [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	if *condition == "" || *conditionVal == "" {
		fmt.Println("you must pass both the condition and condition-val")
		return err
	}

	_, err = cmd.db.UpdateDeviceBy().
		ID(*id).
		Model(*model).
		Name(*name).
		OS(*os).
		Platform(*platform).
		Serial(*serial).
		User(*user).
		UserMDMID(*userid).
		Parse(*condition, *conditionVal).
		Send()

	if err != nil {
		return err
	}

	return err
}
