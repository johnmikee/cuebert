package main

import (
	"flag"
	"fmt"
)

func (c *EC) removeExclusions(args []string) error {
	flagset := flag.NewFlagSet("remove", flag.ExitOnError)
	var (
		email  = flagset.String("email", "", "remove a user row based on the email")
		serial = flagset.String("serial", "", "remove a user row based on the users long name")
	)
	flagset.Usage = usageFor(flagset, "cue exclusions remove [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	base := c.db.RemoveExclusion()
	count := 0

	if *serial != "" {
		base = base.Serial(*serial)
		count++
	}
	if *email != "" {
		base = base.Email(*email)
		count++
	}

	if count > 1 {
		return fmt.Errorf("only one arg can be set for exclusion removal")
	}

	_, err := base.Execute()

	if err != nil {
		return err
	}

	return err
}
