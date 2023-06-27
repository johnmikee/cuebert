package main

import (
	"flag"
	"strconv"
	"strings"

	"github.com/johnmikee/cuebert/db/bot"
)

func (c *BR) getBR(args []string) error {
	flagset := flag.NewFlagSet("get", flag.ExitOnError)
	var (
		all                = flagset.Bool("all", false, "get all bot results")
		serial             = flagset.String("serial", "", "get all bot results for a serial number (comma-separated)")
		userEmail          = flagset.String("user-email", "", "get all bot results for a device owner by email (comma-separated)")
		slackID            = flagset.String("slack-id", "", "get all bot results for a device owner by slack id (comma-separated)")
		managerSlackID     = flagset.String("manager-slack-id", "", "get all bot results for a device owner by manager slack id (comma-separated)")
		firstACKD          = flagset.Bool("first-ackd", false, "get all those who have acknowledged the first message")
		firstMessageSent   = flagset.Bool("first-message-sent", false, "get all those who have received the first message")
		managerMessageSent = flagset.Bool("manager-message-sent", false, "get all those who have received the manager message")
		fullName           = flagset.String("full-name", "", "get all bot results for a device owner by full name (comma-separated)")
		tzOffset           = flagset.String("tz-offset", "", "get all bot results for a device owner by timezone offset (comma-separated)")
	)
	flagset.Usage = usageFor(flagset, "cue br get [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	var br []bot.BotResInfo
	var err error
	if *all {
		br, err = c.db.GetBotTableInfo()
	}

	if *userEmail != "" {
		emailSlice := strings.Split(*userEmail, ",")
		br, err = c.db.UserEmail(emailSlice...)
	}
	if *serial != "" {
		serialSlice := strings.Split(*serial, ",")
		br, err = c.db.GetUsersSerialsBot(serialSlice...)
	}

	if *slackID != "" {
		slackSlice := strings.Split(*slackID, ",")
		br, err = c.db.UserBySlackID(slackSlice...)
	}

	if *managerSlackID != "" {
		managerSlackSlice := strings.Split(*managerSlackID, ",")
		br, err = c.db.UserByManagerSlackID(managerSlackSlice...)
	}

	if *fullName != "" {
		fullNameSlice := strings.Split(*fullName, ",")
		br, err = c.db.UserByFullName(fullNameSlice...)
	}

	if *tzOffset != "" {
		tzOffsetSlice := strings.Split(*tzOffset, ",")
		intSlice := make([]int64, len(tzOffsetSlice))
		for i, v := range tzOffsetSlice {
			sc, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			intSlice[i] = int64(sc)
		}
		br, err = c.db.UserTZOffset(intSlice...)
	}

	if *firstACKD {
		br, err = c.db.GetACKd()
	}

	if *firstMessageSent {
		br, err = c.db.GetFirstMessageSentAll()
	}

	if *managerMessageSent {
		br, err = c.db.ManagerMessageSent()
	}

	if err != nil {
		return err
	}

	brPrinter(br)

	return nil
}
