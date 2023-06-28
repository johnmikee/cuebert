package main

import (
	"flag"
	"os"
	"strings"

	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/internal/env"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Flags holds the args for the program
//
// init is particularly useful when running in a container and you want
// be able to start, stop, and change the config from the bot itself.
type Flags struct {
	authUsers              string // comma separated list of users to perform authorized actions
	authUsersFromIDP       bool   // pull authorized users from the idp. if false use the auth-users flag
	checkInterval          int    // how often to check what cuebert messages need sending
	clearTables            bool   // clear all tables
	cutoffTime             string // cutoffTime will be the time access is revoked
	dailyReport            bool   // send a daily report to the slack channel
	deadline               string // the day the update is required
	deviceDiffInterval     int    // how often to check what devices we need to add/remove
	envType                string // ex: dev, prod
	helpDocsURL            string // url to the help docs
	helpRepoURL            string // url to this repo for the help menu
	helpTicketURL          string // url to the help ticketing system
	idp                    string // ex: okta, onelogin
	init                   bool   // initialize the program and wait for input.
	logLevel               string // ex: debug
	logToFile              bool   // log to file defaults to false
	mdm                    string // ex: jamf, kandji
	pollInterval           int    // how often to poll for reminders
	requiredVers           string // ex: 13.1
	rebuildTablesOnFailure bool   // rebuild tables on an abnormal exit.
	sendManagerMissing     bool   // send a message to the alert channel of missing managers
	serviceName            string // ex: cuebert
	tableNames             string // comma separated list of tables to clear
	testing                bool   // run in testing mode
	testingEndTime         string // the hour the messaging should end
	testingStartTime       string // the hour the messaging should start
	testingUsers           string // comma separated list of users to test with
}

func loadEnv() (*CuebertConfig, []string) {
	f := &Flags{
		authUsers:              "",
		authUsersFromIDP:       true,
		checkInterval:          15,
		clearTables:            true,
		cutoffTime:             "",
		dailyReport:            false,
		deadline:               "",
		deviceDiffInterval:     30,
		envType:                "dev",
		helpDocsURL:            "",
		helpRepoURL:            "https://github.com/johnmikee/cuebert",
		helpTicketURL:          "",
		idp:                    "okta",
		init:                   true,
		logLevel:               "trace",
		logToFile:              false,
		mdm:                    "kandji",
		pollInterval:           10,
		requiredVers:           "13.4.1",
		rebuildTablesOnFailure: false,
		sendManagerMissing:     false,
		serviceName:            "cuebert",
		tableNames:             strings.Join(db.CueTables, ","),
		testing:                true,
		testingEndTime:         "17:00",
		testingStartTime:       "11:00",
		testingUsers:           "",
	}

	flag.BoolVar(
		&f.authUsersFromIDP,
		"auth-users-from-idp",
		f.authUsersFromIDP,
		"Set whether to pull authorized users from the IDP.",
	)
	flag.StringVar(
		&f.authUsers,
		"auth-users",
		f.authUsers,
		"Set which users can perform authorized functions. (comma separated)",
	)
	flag.BoolVar(
		&f.clearTables,
		"clear-tables",
		f.clearTables,
		"Drop all info from tables on initialization.",
	)
	flag.IntVar(
		&f.checkInterval,
		"check-interval",
		f.checkInterval,
		"the number of minutes between device messaging checks and db clean-ups.",
	)
	flag.StringVar(
		&f.cutoffTime,
		"cutoff-time",
		f.cutoffTime,
		"the hour when the install must be done by (HH:MM:SS).",
	)
	flag.StringVar(
		&f.deadline,
		"deadline-date",
		f.deadline,
		"the date the install must be done by (YYYY:MM:DD).",
	)
	flag.IntVar(
		&f.deviceDiffInterval,
		"device-diff-interval",
		f.deviceDiffInterval,
		"the number of minutes between device diff checks.",
	)
	flag.StringVar(
		&f.envType,
		"env-type",
		f.envType,
		"Set the env type. Options are [prod, dev].",
	)
	flag.StringVar(
		&f.helpDocsURL,
		"help-docs-url",
		f.helpDocsURL,
		"the url to the cuebert docs.",
	)
	flag.StringVar(
		&f.helpRepoURL,
		"help-repo-url",
		f.helpRepoURL,
		"the url to the cuebert repo.",
	)
	flag.StringVar(
		&f.helpTicketURL,
		"help-ticket-url",
		f.helpTicketURL,
		"the url to the cuebert ticketing system.",
	)
	flag.StringVar(
		&f.idp,
		"idp",
		f.idp,
		"Set the IDP to use. Options are [okta].",
	)
	flag.BoolVar(
		&f.init,
		"init",
		f.init,
		"Start the program, load the config, and wait for input before running.",
	)
	flag.BoolVar(
		&f.logToFile,
		"log-to-file",
		f.logToFile,
		"Log results to file.",
	)
	flag.StringVar(
		&f.logLevel,
		"log-level",
		f.logLevel,
		"Set the log level.",
	)
	flag.StringVar(
		&f.mdm,
		"mdm",
		f.mdm,
		"Set the MDM to use. Options are [jamf, kandji].",
	)
	flag.BoolVar(
		&f.rebuildTablesOnFailure,
		"rebuild-tables-on-failure",
		f.rebuildTablesOnFailure,
		"rebuild tables on an abnormal exit.",
	)
	flag.BoolVar(
		&f.sendManagerMissing,
		"send-manager-missing",
		f.sendManagerMissing,
		"send a message to the alert channel of missing managers",
	)
	flag.StringVar(
		&f.serviceName,
		"service-name",
		f.serviceName,
		"if using the dev env the service name to store keys under.",
	)
	flag.StringVar(
		&f.tableNames,
		"table-names",
		f.tableNames,
		"a list of tables to clear on initialization. (comma separated)",
	)
	flag.BoolVar(
		&f.testing,
		"testing",
		f.testing,
		"Log actions that would take place instead of performing them.",
	)
	flag.StringVar(
		&f.testingEndTime,
		"testing-end-time",
		f.testingEndTime,
		"the time to end testing (HH:MM).",
	)
	flag.StringVar(
		&f.testingStartTime,
		"testing-start-time",
		f.testingStartTime,
		"the time to start testing (HH:MM).",
	)
	flag.StringVar(
		&f.testingUsers,
		"testing-users",
		f.testingUsers,
		"a list of slack id's to perform the actions on during testing instead of every user. (comma separated)",
	)
	flag.StringVar(
		&f.requiredVers,
		"required-os",
		f.requiredVers,
		"the version to require for the fleet",
	)
	flag.IntVar(
		&f.pollInterval,
		"poll-interval",
		f.pollInterval,
		"the number of minutes between device polling checks.",
	)
	flag.BoolVar(
		&f.dailyReport,
		"daily-report",
		f.dailyReport,
		"send a daily report to the admin alert channel.",
	)

	flag.Parse()

	envArgs := os.Args[0:]

	log := logger.NewLogger(
		&logger.Config{
			ToFile:  f.logToFile,
			Level:   f.logLevel,
			Service: f.serviceName,
			Env:     f.envType,
		},
	)

	var cfg CuebertConfig
	err := env.Get(
		env.EnvType(f.envType),
		&env.GetConfig{
			Name:         f.serviceName,
			EnvPrefix:    "CUEBERT",
			ConfigStruct: &cfg,
			Type:         env.JSON,
		},
	)
	if err != nil {
		log.Info().Msg("no config found, exiting")
		os.Exit(1)
	}

	cfg.flags = f
	cfg.log = log

	if f.testing {
		if f.testingUsers != "" {
			userSlice := strings.Split(f.testingUsers, ",")
			cfg.testUsers = append(cfg.testUsers, userSlice...)
		}
	}

	return &cfg, envArgs
}

func (b *Bot) logFlags() {
	b.log.Trace().
		Str("authUsers", b.cfg.flags.authUsers).
		Bool("authUsersFromIDP", b.cfg.flags.authUsersFromIDP).
		Int("checkInterval", b.cfg.flags.checkInterval).
		Bool("clearTables", b.cfg.flags.clearTables).
		Str("cutoffTime", b.cfg.flags.cutoffTime).
		Bool("dailyReport", b.cfg.flags.dailyReport).
		Str("deadline", b.cfg.flags.deadline).
		Int("deviceDiffInterval", b.cfg.flags.deviceDiffInterval).
		Str("envType", b.cfg.flags.envType).
		Str("helpDocsURL", b.cfg.flags.helpDocsURL).
		Str("helpRepoURL", b.cfg.flags.helpRepoURL).
		Str("helpTicketURL", b.cfg.flags.helpTicketURL).
		Str("idp", b.cfg.flags.idp).
		Bool("init", b.cfg.flags.init).
		Str("tableNames", b.cfg.flags.tableNames).
		Str("logLevel", b.cfg.flags.logLevel).
		Bool("logToFile", b.cfg.flags.logToFile).
		Str("mdm", b.cfg.flags.mdm).
		Int("pollInterval", b.cfg.flags.pollInterval).
		Str("requiredVersion", b.cfg.flags.requiredVers).
		Bool("rebuildTablesOnFailure", b.cfg.flags.rebuildTablesOnFailure).
		Bool("sendManagerMissing", b.cfg.flags.sendManagerMissing).
		Str("serviceName", b.cfg.flags.serviceName).
		Str("tableNames", b.cfg.flags.tableNames).
		Bool("testing", b.cfg.flags.testing).
		Str("testingEndTime", b.cfg.flags.testingEndTime).
		Str("testingStartTime", b.cfg.flags.testingStartTime).
		Str("testingUsers", b.cfg.flags.testingUsers).
		Msg("current configuration")
}
