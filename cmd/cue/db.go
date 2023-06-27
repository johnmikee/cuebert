package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/internal/device"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type DB struct {
	db      *db.Config
	devices *device.Device
	log     logger.Logger
	run     func([]string) error
}

func (cmd *DB) usage() {
	const help = `cue db:

Valid Options:
  * info
  * init
  * test-connection

Use cue db <option> -h for additional usage of each command.
Example: cue db info -h
	`
	fmt.Print(help)
}

func (c *CueConfig) DB(args []string) error {
	cmd := &DB{
		devices: c.devices,
		db:      c.db,
		log:     c.log,
	}
	if len(args) < 1 {
		cmd.usage()
		os.Exit(1)
	}

	switch strings.ToLower(args[0]) {
	case "info":
		cmd.run = cmd.dbInfo
	case "init":
		cmd.run = cmd.dbInit
	case "test-connection":
		cmd.run = cmd.dbTestConn
	default:
		cmd.usage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func (cmd *DB) infousage() {
	const help = `cue db info:

Valid Options:
  * config

Use cue db info <option> -h for additional usage of each command.
Example: cue db info config -h
	`
	fmt.Print(help)
}

func (cmd *DB) dbInfo(args []string) error {
	if len(args) < 1 {
		cmd.infousage()
		os.Exit(1)
	}
	switch strings.ToLower(args[0]) {
	case "config":
		cmd.run = cmd.dbConfig
	default:
		cmd.infousage()
		os.Exit(1)
	}

	return cmd.run(args[1:])
}

func (cmd *DB) dbTestConn(args []string) error {
	err := cmd.db.TestConnection()
	if err != nil {
		msg := fmt.Sprintf("Error connecting to the DB: %v", err)
		fmt.Printf("\033[31m" + msg + "\033[0m" + "\n")
		return err
	}
	fmt.Println("\033[32m" + "Successful Connection" + "\033[0m")
	return nil
}
