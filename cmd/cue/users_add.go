package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func (cmd *UC) addusage() {
	const help = `cue users add:

Valid Options:
  * all
  * some

Use cue users add <option> -h for additional usage of each command.
Example: cue users add some -h
	`
	fmt.Print(help)
}

func (cmd *UC) addUsers(args []string) error {
	if len(args) < 1 {
		cmd.addusage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "all":
		err := cmd.addAllUsers()
		cmd.run = func(s []string) error { return err }

	case "some":
		cmd.run = cmd.addSome

	default:
		cmd.addusage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func (c *UC) addAllUsers() error {
	_, err := c.users.AddAllUsers()
	return err
}

func (cmd *UC) addSome(args []string) error {
	var err error
	flagset := flag.NewFlagSet("add", flag.ExitOnError)
	var (
		id       = flagset.String("id", "", "user id")
		longName = flagset.String("long-name", "", "user full name")
		email    = flagset.String("email", "", "user email")
		slackID  = flagset.String("slack-id", "", "user slack ID")
	)
	flagset.Usage = usageFor(flagset, "cue users add some [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	_, err = cmd.db.AddUser().
		ID(*id).
		Email(*email).
		Slack(*slackID).
		LongName(*longName).
		Execute()

	return err
}
