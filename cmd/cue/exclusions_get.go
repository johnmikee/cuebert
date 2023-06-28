package main

import (
	"flag"
	"strings"

	"github.com/johnmikee/cuebert/db/exclusions"
)

func (cmd *EC) getExclusions(args []string) error {
	flagset := flag.NewFlagSet("get", flag.ExitOnError)
	var (
		all       = flagset.Bool("all", false, "get all exclusions")
		serial    = flagset.String("serial", "", "get all exclusions for a serial number (comma-separated)")
		userEmail = flagset.String("user-email", "", "get all exclusions for a device owner by email (comma-separated)")
	)
	flagset.Usage = usageFor(flagset, "cue exclusions get [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	var ei []exclusions.ExclusionInfo
	var err error
	if *all {
		ei, err = cmd.db.ExclusionBy().All().Query()
	}

	if *serial != "" {
		serialSlice := strings.Split(*serial, ",")
		ei, err = cmd.db.ExclusionBy().Serial(serialSlice...).Query()
	}

	if *userEmail != "" {
		emailSlice := strings.Split(*userEmail, ",")
		ei, err = cmd.db.ExclusionBy().Email(emailSlice...).Query()
	}

	if err != nil {
		return err
	}

	exclusionPrinter(ei)

	return nil
}
