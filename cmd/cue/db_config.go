package main

import (
	"flag"
	"fmt"

	"github.com/johnmikee/cuebert/internal/db"
)

func (cmd *DB) dbConfig(args []string) error {
	flagset := flag.NewFlagSet("config", flag.ExitOnError)
	var (
		host     = flagset.Bool("host", false, "address of the db")
		name     = flagset.Bool("name", false, "name of the db")
		password = flagset.Bool("password", false, "password used to access the db")
		port     = flagset.Bool("port", false, "port the db is running on")
		user     = flagset.Bool("user", false, "user accessing the db")
	)
	flagset.Usage = usageFor(flagset, "cue db info config [flags]")

	if err := flagset.Parse(args); err != nil {
		return err
	}

	if *host {
		fmt.Printf("host: %s\n", cmd.db.Print(db.Host))
	}

	if *name {
		fmt.Printf("name: %s\n", cmd.db.Print(db.Name))
	}

	if *password {
		fmt.Printf("password: %s\n", cmd.db.Print(db.Password))
	}

	if *port {
		fmt.Printf("port: %s\n", cmd.db.Print(db.Port))
	}

	if *user {
		fmt.Printf("user: %s\n", cmd.db.Print(db.UserName))
	}

	return nil
}
