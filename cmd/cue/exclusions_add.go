package main

import (
	"flag"
	"fmt"
	"time"
)

func (cmd *EC) add(args []string) error {
	var err error
	flagset := flag.NewFlagSet("add", flag.ExitOnError)
	var (
		serial    = flagset.String("serial", "", "device serial number")
		userEmail = flagset.String("user-email", "", "device owners email")
		reason    = flagset.String("reason", "", "explanation for exclusion")
		until     = flagset.String("until", "", "date for the exclusion to last until (YYYY-MM-DD)")
	)
	flagset.Usage = usageFor(flagset, "cue exclusions add [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	if *serial == "" || *userEmail == "" {
		return fmt.Errorf("must provided both a serial number and email to add an exclusion")
	}

	ts, err := time.Parse("2006-01-02", *until)
	if err != nil {
		return err
	}
	_, err = cmd.db.AddExclusions().
		Email(*userEmail).
		SerialNumber(*serial).
		Reason(*reason).
		Until(ts).
		Execute()

	if err != nil {
		return err
	}

	return err
}
