package main

import (
	"flag"
	"os"
	"strings"

	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"

	"github.com/johnmikee/cuebert/pkg/env"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/johnmikee/yae"
)

func loadEnv() (*Cuebert, []string) {
	f := &config.Flags{
		AuthorizedUsers:         "",
		AuthUsersFromIDP:        true,
		CheckInterval:           30,
		CutoffTime:              "5:00PM",
		Deadline:                "04/12",
		Debug:                   false,
		DefaultReminderInterval: 60,
		DiffInterval:            15,
		EnvType:                 string(env.PROD),
		HelpDocsURL:             "",
		HelpRepoURL:             "https://github.com/johnmikee/cuebert",
		HelpTicketURL:           "",
		IDP:                     "",
		Init:                    true,
		LogLevel:                string(logger.TRACE),
		LogToFile:               true,
		MakeMigrations:          false,
		MDM:                     "",
		PollInterval:            10,
		RequiredVers:            "14.4.1",
		ServiceName:             "cuebert",
		Task:                    string(task.Update),
		TaskID:                  "",
		TaskImplementation:      string(implementation.OS),
		Testing:                 false,
		TestingEndTime:          "18:00",
		TestingStartTime:        "11:00",
		TestingUsers:            "",
		TestUsers:               []string{},
	}
	var testDBConn bool

	flag.BoolVar(
		&f.AuthUsersFromIDP,
		"auth-users-from-idp",
		f.AuthUsersFromIDP,
		"Set whether to pull authorized users from the IDP.",
	)
	flag.StringVar(
		&f.AuthorizedUsers,
		"auth-users",
		f.AuthorizedUsers,
		"Set which users can perform authorized functions. (comma separated)",
	)
	flag.IntVar(
		&f.CheckInterval,
		"check-interval",
		f.CheckInterval,
		"the number of minutes between device messaging checks and db clean-ups.",
	)
	flag.StringVar(
		&f.CutoffTime,
		"cutoff-time",
		f.CutoffTime,
		"the hour when the install must be done by (HH:MM:SS).",
	)
	flag.StringVar(
		&f.Deadline,
		"deadline-date",
		f.Deadline,
		"the date the install must be done by (YYYY:MM:DD).",
	)
	flag.BoolVar(
		&f.Debug,
		"debug",
		f.Debug,
		"Run the slacker client in Debug mode.",
	)
	flag.IntVar(
		&f.DefaultReminderInterval,
		"default-reminder-interval",
		f.DefaultReminderInterval,
		"the number of minutes between reminders.",
	)
	flag.IntVar(
		&f.DiffInterval,
		"diff-interval",
		f.DiffInterval,
		"the number of minutes between diff checks.",
	)
	flag.StringVar(
		&f.EnvType,
		"env-type",
		f.EnvType,
		"Set the env type. Options are [prod, dev].",
	)
	flag.StringVar(
		&f.HelpDocsURL,
		"help-docs-url",
		f.HelpDocsURL,
		"the url to the cuebert docs.",
	)
	flag.StringVar(
		&f.HelpRepoURL,
		"help-repo-url",
		f.HelpRepoURL,
		"the url to the cuebert repo.",
	)
	flag.StringVar(
		&f.HelpTicketURL,
		"help-ticket-url",
		f.HelpTicketURL,
		"the url to the cuebert ticketing system.",
	)
	flag.StringVar(
		&f.IDP,
		"idp",
		f.IDP,
		"Set the IDP to use. Options are [okta].",
	)
	flag.BoolVar(
		&f.Init,
		"init",
		f.Init,
		"Start the program, load the config, and wait for input before running.",
	)
	flag.BoolVar(
		&f.LogToFile,
		"log-to-file",
		f.LogToFile,
		"Log results to file.",
	)
	flag.StringVar(
		&f.LogLevel,
		"log-level",
		f.LogLevel,
		"Set the log level.",
	)
	flag.BoolVar(
		&f.MakeMigrations,
		"make-migrations",
		f.MakeMigrations,
		"Create the migrations and exit.",
	)
	flag.StringVar(
		&f.MDM,
		"mdm",
		f.MDM,
		"Set the MDM to use. Options are [jamf, kandji].",
	)
	flag.StringVar(
		&f.ServiceName,
		"service-name",
		f.ServiceName,
		"if using the dev env the service name to store keys under.",
	)
	flag.StringVar(
		&f.Task,
		"task",
		f.Task,
		"what task is being preformed (ex: updates or compliance)",
	)
	flag.StringVar(
		&f.TaskImplementation,
		"task-impl",
		f.TaskImplementation,
		"if using the compliance task, what is carrying that out",
	)
	flag.StringVar(
		&f.TaskID,
		"task-id",
		f.TaskID,
		"if using the compliance task, what is the id of the task",
	)
	flag.BoolVar(
		&f.Testing,
		"testing",
		f.Testing,
		"Log actions that would take place instead of performing them.",
	)
	flag.StringVar(
		&f.TestingEndTime,
		"testing-end-time",
		f.TestingEndTime,
		"the time to end testing (HH:MM).",
	)
	flag.StringVar(
		&f.TestingStartTime,
		"testing-start-time",
		f.TestingStartTime,
		"the time to start testing (HH:MM).",
	)
	flag.StringVar(
		&f.TestingUsers,
		"testing-users",
		f.TestingUsers,
		"a list of slack id's to perform the actions on during testing instead of every user. (comma separated)",
	)
	flag.StringVar(
		&f.RequiredVers,
		"required-os",
		f.RequiredVers,
		"the version to require for the fleet",
	)
	flag.IntVar(
		&f.PollInterval,
		"poll-interval",
		f.PollInterval,
		"the number of minutes between device polling checks.",
	)
	flag.BoolVar(
		&testDBConn,
		"test-db-conn",
		false,
		"test the db connection and exit.",
	)
	flag.Parse()

	envArgs := os.Args[0:]

	log := logger.New(
		&logger.Config{
			ToFile:  f.LogToFile,
			Level:   f.LogLevel,
			Service: f.ServiceName,
			Env:     f.EnvType,
		},
	)

	var cfg config.Config
	err := yae.Get(
		yae.EnvType(f.EnvType),
		&yae.Env{
			EnvPrefix:    "CUEBERT",
			ConfigStruct: &cfg,
			Type:         yae.JSON,
		},
	)
	if err != nil {
		log.Info().AnErr("error", err).Msg("no config found, exiting")
		os.Exit(1)
	}

	config := &Cuebert{
		flags:  f,
		log:    log,
		config: &cfg,
	}

	if f.Testing {
		if f.TestingUsers != "" {
			userSlice := strings.Split(f.TestingUsers, ",")
			config.flags.TestUsers = append(config.flags.TestUsers, userSlice...)
		}
	}

	if testDBConn {
		_, _ = dbConn(config, true)
	}
	return config, envArgs
}
