package main

import (
	"flag"
	"fmt"
)

func (c *CueConfig) removeUsers(args []string) error {
	var err error
	flagset := flag.NewFlagSet("remove", flag.ExitOnError)
	var (
		all      = flagset.Bool("all", false, "remove all users")
		id       = flagset.String("id", "", "remove a user row based on the users id")
		email    = flagset.String("email", "", "remove a user row based on the email")
		longName = flagset.String("name", "", "remove a user row based on the users long name")
		slackID  = flagset.String("slack-id", "", "remove a user row based on the users slack id")
	)
	flagset.Usage = usageFor(flagset, "cue users remove [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	base := c.db.RemoveUserBy()
	count := 0

	if *all {
		base.All()
		count++
	}

	if *id != "" {
		base.ID(*id)
		count++
	}
	if *email != "" {
		base.Email(*email)
		count++
	}
	if *longName != "" {
		base.UserLongName(*longName)
		count++
	}
	if *slackID != "" {
		base.SlackID(*slackID)
		count++
	}

	if count > 1 {
		return fmt.Errorf("only one arg can be set for user removal")
	}

	_, err = base.Run()

	return err
}
