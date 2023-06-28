package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/johnmikee/cuebert/db/users"
	"github.com/johnmikee/cuebert/mdm"
)

func (cmd *UC) sourceusage() {
	const help = `cue users get:

Valid Options:
  * db
  * mdm

Use cue users get [option] -h for additional usage of each command.
Example: cue users get db -h
	`
	fmt.Print(help)
}

func (cmd *UC) getUserSource(args []string) error {
	var err error
	if len(args) < 1 {
		cmd.sourceusage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "db":
		err = cmd.c.getUsersDB(args[1:])
	case "mdm":
		err = cmd.getUsersMDM(args[1:])

	default:
		cmd.sourceusage()
		os.Exit(1)
	}

	return err
}

func (c *CueConfig) getUsersDB(args []string) error {
	flagset := flag.NewFlagSet("db", flag.ExitOnError)
	var (
		all      = flagset.Bool("all", false, "return all users")
		id       = flagset.String("id", "", "returns users based on the user id (comma-separated)")
		longName = flagset.String("long-name", "", "returns users based on the users full name (comma-separated)")
		email    = flagset.String("email", "", "returns users based on the users email (comma-separated)")
		slackID  = flagset.String("slack-id", "", "returns users based on the users slack id (comma-separated)")
	)
	flagset.Usage = usageFor(flagset, "cue users get db [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	var err error
	var ui []users.UserInfo
	if *all {
		ui, err = c.db.UserBy().All().Query()
	}

	if *id != "" {
		idSlice := strings.Split(*id, ",")
		ui, err = c.db.UserBy().ID(idSlice...).Query()
	}

	if *longName != "" {
		lnSlice := strings.Split(*longName, ",")
		ui, err = c.db.UserBy().LongName(lnSlice...).Query()
	}

	if *email != "" {
		emailSlice := strings.Split(*email, ",")
		ui, err = c.db.UserBy().Email(emailSlice...).Query()
	}

	if *slackID != "" {
		slackSlice := strings.Split(*slackID, ",")
		ui, err = c.db.UserBy().SlackID(slackSlice...).Query()
	}

	if err != nil {
		return err
	}

	userDBPrinter(ui)

	return nil
}

func (c *UC) getUsersMDM(args []string) error {
	flagset := flag.NewFlagSet("users", flag.ExitOnError)
	var (
		all    = flagset.Bool("all", false, "return all users")
		emails = flagset.String("emails", "", "query by user email (comma-separated)")
		ids    = flagset.String("ids", "", "query by user id (comma-separated)")
		names  = flagset.String("names", "", "query by user name in mdm (comma-separated)")
	)
	flagset.Usage = usageFor(flagset, "cue get users [flags]")
	if err := flagset.Parse(args); err != nil {
		return err
	}

	var mu []mdm.User
	var err error

	if *all {
		mu, err = c.users.GetMDMUsers(nil)
	}
	if *emails != "" || *ids != "" || *names != "" {
		mu, err = c.users.GetMDMUsers(
			&mdm.QueryOpts{
				UserEmail: *emails,
				UserID:    *ids,
				UserName:  *names,
			},
		)
	}

	if err != nil {
		return err
	}

	userMDMPrinter(mu)

	return err
}
