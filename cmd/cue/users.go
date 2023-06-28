package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/johnmikee/cuebert/db/users"
	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/internal/user"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type UC struct {
	c     *CueConfig
	db    *db.Config
	log   logger.Logger
	users *user.User
	run   func([]string) error
}

func (out *TableOutput) brHeader() {
	fmt.Fprintf(out.w, "slack_id\tuser_email\tmanager_slack_id\tfirst_ack\tfirst_ack_time\tfirst_message_sent\tfirst_message_sent_at\tmanager_message_sent\tmanager_message_sent_at\tfull_name\tdelay_at\tdelay_date\tdelay_time\tdelay_sent\tserial_number\ttz_offset\tcreated_at\tupdated_at\n")
}
func (out *TableOutput) exclusionsHeader() {
	fmt.Fprintf(out.w, "approved\tserial_number\tuser_email\treason\tuntil\tcreated_at\tupdated_at\n")
}

func (out *TableOutput) usersHeader() {
	fmt.Fprintf(out.w, "user_mdm_id\tuser_long_name\tuser_email\tuser_slack_id\tcreated_at\tupdated_at\n")
}

func (out *TableOutput) usersMDMHeader() {
	fmt.Fprintf(out.w, "email\tname\tid\n")
}

func (cmd *UC) usage() {
	const help = `cue users:

Valid Options:
  * add
  * get
  * remove

Use cue users <option> -h for additional usage of each command.
Example: cue users get -h
	`
	fmt.Print(help)
}

func (c *CueConfig) Users(args []string) error {
	cmd := &UC{
		c:     c,
		db:    c.db,
		log:   c.log,
		users: c.users,
	}
	if len(args) < 1 {
		cmd.usage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "add":
		cmd.run = cmd.addUsers

	case "get":
		cmd.run = cmd.getUserSource

	case "remove":
		cmd.run = cmd.c.removeUsers

	// case "update":
	// 	cmd.run = cmd.updateDevices
	default:
		cmd.usage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func userDBPrinter(users []users.UserInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 6, 2, ' ', 0)
	out := &TableOutput{w}
	out.usersHeader()
	defer out.basicFooter()
	for _, u := range users {
		fmt.Fprintf(out.w, "%s\t%s\t%s\t%ss\t%s\t%s\n",
			u.MDMID,
			u.UserLongName,
			u.UserEmail,
			u.UserSlackID,
			u.CreatedAt,
			u.UpdatedAt)
	}
}

func userMDMPrinter(users []mdm.User) {
	w := tabwriter.NewWriter(os.Stdout, 0, 6, 2, ' ', 0)
	out := &TableOutput{w}
	out.usersMDMHeader()
	defer out.basicFooter()
	for _, u := range users {
		fmt.Fprintf(out.w, "%s\t%s\t%s\n",
			u.Email,
			u.Name,
			u.ID)
	}
}
