package main

import (
	"flag"
	"strings"

	"github.com/johnmikee/cuebert/db"
)

func (c *BR) removeBR(args []string) error {
	flagset := flag.NewFlagSet("remove", flag.ExitOnError)
	var (
		all            = flagset.Bool("all", false, "remove all bot results")
		email          = flagset.String("email", "", "remove a bot_result row based on the email")
		serial         = flagset.String("serial", "", "remove a bot_result row based on the users long name")
		slackID        = flagset.String("slack-id", "", "remove a bot_result row based on the users slack id")
		managerSlackID = flagset.String("manager-slack-id", "", "remove a bot_result row based on the users manager slack id")
		fullName       = flagset.String("full-name", "", "remove a bot_result row based on the users full name")
	)
	flagset.Usage = usageFor(flagset, "cue br remove [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	if *all {
		return c.db.Delete(db.CueTables)
	}
	if *email != "" {
		emailSlice := strings.Split(*email, ",")
		_, err := c.db.RemoveBRBy().UserEmail(emailSlice...).Execute()
		return err
	}

	if *serial != "" {
		serialSlice := strings.Split(*serial, ",")
		_, err := c.db.RemoveBRBy().Serial(serialSlice...).Execute()
		return err
	}

	if *slackID != "" {
		slackSlice := strings.Split(*slackID, ",")
		_, err := c.db.RemoveBRBy().SlackID(slackSlice...).Execute()
		return err
	}

	if *managerSlackID != "" {
		managerSlice := strings.Split(*managerSlackID, ",")
		_, err := c.db.RemoveBRBy().ManagerSlackID(managerSlice...).Execute()
		return err
	}

	if *fullName != "" {
		fullNameSlice := strings.Split(*fullName, ",")
		_, err := c.db.RemoveBRBy().FullName(fullNameSlice...).Execute()
		return err
	}

	return nil
}
