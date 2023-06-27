package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type BR struct {
	c   *Config
	db  *db.Config
	log logger.Logger
	run func([]string) error
}

func (cmd *BR) usage() {
	const help = `cue br:

Valid Options:
  * get
  * remove

Use cue br <option> -h for additional usage of each command.
Example: cue br get -h
	`
	fmt.Print(help)
}

func (c *CueConfig) BotRes(args []string) error {
	cmd := &BR{
		c:   c.conf,
		db:  c.db,
		log: c.log,
	}
	if len(args) < 1 {
		cmd.usage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "get":
		cmd.run = cmd.getBR
	case "remove":
		cmd.run = cmd.removeBR
	default:
		cmd.usage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func brPrinter(res []bot.BotResInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 75, 2, ' ', 0)
	out := &TableOutput{w}
	out.brHeader()
	defer out.basicFooter()
	for i := range res {
		fmt.Fprintf(out.w, "%s\t%s\t%s\t%t\t%v\t%t\t%v\t%t\t%v\t%s\t%v\t%s\t%s\t%t\t%s\t%v\t%v\t%v\n",
			res[i].SlackID,
			res[i].UserEmail,
			res[i].ManagerSlackID,
			res[i].FirstACK,
			res[i].FirstACKTime,
			res[i].FirstMessageSent,
			res[i].FirstMessageSentAt,
			res[i].ManagerMessageSent,
			res[i].ManagerMessageSentAt,
			res[i].FullName,
			res[i].DelayAt,
			res[i].DelayDate,
			res[i].DelayTime,
			res[i].DelaySent,
			res[i].SerialNumber,
			res[i].TZOffset,
			res[i].CreatedAt,
			res[i].UpdatedAt,
		)
	}
}
