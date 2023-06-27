package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func (cmd *DB) initusage() {
	const help = `cue db init:

Valid Options:
  * schema
  * values

Use cue db init <option> -h for additional usage of each command.
Example: cue db init values -h
	`
	fmt.Print(help)
}

func (cmd *DB) dbInit(args []string) error {
	if len(args) < 1 {
		cmd.initusage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "schema":
		cmd.run = cmd.dbSchema
	case "values":
		cmd.run = cmd.initTableValues
	default:
		cmd.initusage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func (cmd *DB) initTableValues(args []string) error {
	flagset := flag.NewFlagSet("init", flag.ExitOnError)

	var (
		devices = flagset.Bool("devices", false, "add values to the devices table")
		users   = flagset.Bool("users", false, "remove values to the users table")
	)

	flagset.Usage = usageFor(flagset, "cue db values init [flags]")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	if *devices {
		err := cmd.devices.AddAllDevices()
		if err != nil {
			return err
		}
		cmd.log.Info().Msg("devices table has been successfully populated")
	}

	if *users {
		// err := cmd.addAllUsers()
		// if err != nil {
		// 	return err
		// }
		cmd.log.Info().Msg("users table has been successfully populated")
	}

	return nil
}

func (cmd *DB) dbSchema(args []string) error {
	flagset := flag.NewFlagSet("schema", flag.ExitOnError)

	var (
		all    = flagset.Bool("all", false, "run all options below")
		db     = flagset.Bool("db", false, "create the cue db")
		user   = flagset.Bool("name", false, "create the cue user")
		tables = flagset.Bool("password", false, "create the tables on the cue db")
	)

	flagset.Usage = usageFor(flagset, "cue db init schema [flags]")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	initScript := "resources/db/create.sh"

	base, err := shellCMD("git", "rev-parse", "--show-toplevel")
	if err != nil {
		fmt.Printf("error getting top level directory %s", err.Error())
		return fmt.Errorf("error getting top level directory %s", err.Error())
	}

	scrPath := fmt.Sprintf("%s/%s", strings.TrimSpace(base), initScript)

	if _, err := os.Stat(scrPath); !os.IsNotExist(err) {
		out, err := shellCMD(
			"bash", scrPath,
			"-a", strconv.FormatBool(*all),
			"-d", strconv.FormatBool(*db),
			"-t", strconv.FormatBool(*tables),
			"-u", strconv.FormatBool(*user),
		)
		if err != nil {
			return fmt.Errorf("error running db init script %s", err.Error())
		}
		fmt.Printf("result: %s\n", out)
	}
	return nil
}
