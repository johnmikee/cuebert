package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/johnmikee/cuebert/db/exclusions"
	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/internal/device"
	"github.com/johnmikee/cuebert/internal/user"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type EC struct {
	devices *device.Device
	users   *user.User
	log     logger.Logger
	db      *db.Config
	run     func([]string) error
}

func (cmd *EC) usage() {
	const help = `cue exclusions:

Valid Options:
  * add
  * get
  * remove

Use cue exclusions <option> -h for additional usage of each command.
Example: cue exclusions get -h
	`
	fmt.Print(help)
}

func (c *CueConfig) Exclusions(args []string) error {
	cmd := &EC{
		devices: c.devices,
		users:   c.users,
		log:     c.log,
		db:      c.db,
	}

	if len(args) < 1 {
		cmd.usage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "add":
		cmd.run = cmd.add
	case "get":
		cmd.run = cmd.getExclusions
	case "remove":
		cmd.run = cmd.removeExclusions
	default:
		cmd.usage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func exclusionPrinter(users []exclusions.ExclusionInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 75, 2, ' ', 0)
	out := &TableOutput{w}
	out.exclusionsHeader()
	defer out.basicFooter()
	for i := range users {
		fmt.Fprintf(out.w, "%v\t%s\t%s\t%s\t%ss\t%s\n",
			users[i].Approved,
			users[i].SerialNumber,
			users[i].UserEmail,
			users[i].Reason,
			users[i].CreatedAt,
			users[i].UpdatedAt)
	}
}
