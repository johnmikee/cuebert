package main

import (
	"flag"
	"os"
	"strings"

	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/device"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/cuebert/method"
	mc "github.com/johnmikee/cuebert/cuebert/method/config"
	"github.com/johnmikee/cuebert/cuebert/tables"
	"github.com/johnmikee/cuebert/cuebert/user"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/idp"
	ic "github.com/johnmikee/cuebert/idp/client"
	"github.com/johnmikee/cuebert/mdm"

	mdmclient "github.com/johnmikee/cuebert/mdm/client"
	"github.com/johnmikee/cuebert/pkg/env"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/slack-go/slack"
)

func loadEnv() (*Cuebert, []string) {
	f := &Flags{
		authUsers:               "",
		authUsersFromIDP:        true,
		checkInterval:           15,
		clearTables:             true,
		cutoffTime:              "",
		dailyReport:             false,
		deadline:                "",
		defaultReminderInterval: 60,
		deviceDiffInterval:      30,
		envType:                 "dev",
		helpDocsURL:             "https://help.megacorp.com/cuebert",
		helpRepoURL:             "https://github.com/johnmikee/cuebert",
		helpTicketURL:           "https://tickets.megacorp.com/cuebert",
		idp:                     "okta",
		init:                    true,
		logLevel:                "trace",
		logToFile:               false,
		mdm:                     "kandji",
		method:                  "manager",
		pollInterval:            10,
		requiredVers:            "13.4.1",
		rebuildTablesOnFailure:  false,
		sendManagerMissing:      false,
		serviceName:             "cuebert",
		tableNames:              strings.Join(db.CueTables, ","),
		testing:                 true,
		testingEndTime:          "17:00",
		testingStartTime:        "11:00",
		testingUsers:            "",
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
		&f.defaultReminderInterval,
		"default-reminder-interval",
		f.defaultReminderInterval,
		"the number of minutes between reminders.",
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
	flag.StringVar(
		&f.method,
		"method",
		f.method,
		"Set the method to use. Options are [manager, device].",
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

	var cfg Config
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

	config := &Cuebert{
		flags:  f,
		log:    log,
		config: &cfg,
	}

	if f.testing {
		if f.testingUsers != "" {
			userSlice := strings.Split(f.testingUsers, ",")
			config.testUsers = append(config.testUsers, userSlice...)
		}
	}

	return config, envArgs
}

func setup() *Cuebert {
	cb, _ := loadEnv()

	conn, err := tables.Connect(
		&tables.Conf{
			Host:     cb.config.DBAddress,
			Name:     cb.config.DBName,
			Password: cb.config.DBPass,
			Port:     cb.config.DBPort,
			User:     cb.config.DBUser,
		},
	)

	if err != nil {
		cb.log.Info().AnErr("connecting to db", err).Send()
		os.Exit(3)
	}
	cb.db = conn.DB

	idpclient := ic.New(
		&ic.IDP{
			IDP: idp.IDP(cb.flags.idp),
			Config: idp.Config{
				Domain: cb.config.IDPDomain,
				URL:    cb.config.IDPURL,
				Token:  cb.config.IDPToken,
				Client: nil,
				Log:    cb.log,
			},
		},
	)
	mdmclient := mdmclient.New(
		&mdmclient.MDM{
			MDM: mdm.MDM(cb.flags.mdm),
			Config: mdm.Config{
				Domain: cb.config.IDPDomain,
				MDM:    mdm.MDM(cb.flags.mdm),
				URL:    cb.config.MDMURL,
				Token:  cb.config.MDMKey,
				Client: nil,
				Log:    cb.log,
			},
		},
	)
	tables := tables.New(
		conn.DB,
		&cb.log,
		tables.WithDevices(
			device.New(
				&device.Config{
					Client: mdmclient,
					DB:     conn.DB,
					Log:    &cb.log,
				},
			),
		),
		tables.WithUsers(
			user.New(
				&user.Config{
					DB:     conn.DB,
					Log:    &cb.log,
					Slack:  slack.New(cb.config.SlackBotToken),
					Client: mdmclient,
				},
			),
		),
	)
	method := mc.New(
		&mc.Method{
			Method: method.Option(cb.flags.method),
			Config: method.Config{
				Log:               cb.log,
				Tables:            tables,
				Bot:               cb.bot,
				StatusHandler:     cb.statusHandler,
				SlackClient:       slack.New(cb.config.SlackBotToken),
				IDP:               idpclient,
				MDM:               mdmclient,
				SlackAlertChannel: cb.config.SlackAlertChannel,
				CutoffTime:        cb.flags.cutoffTime,
				Deadline:          cb.flags.deadline,
				RequiredVers:      cb.flags.requiredVers,
				Testing:           cb.flags.testing,
				TestingUsers:      cb.testUsers,
				PollInterval:      cb.flags.pollInterval,
			},
		},
	)
	cb.tables = tables
	cb.idp = idpclient
	cb.mdm = mdmclient
	cb.method = method
	cb.reloadSignal = make(chan struct{})
	cb.statusChan = make(chan handlers.StatusMessage)
	cb.statusHandler = &handlers.StatusHandler{}
	cb.startSignal = make(chan struct{})
	cb.stopSignal = make(chan struct{})
	cb.isRunning = false

	cb.bot = bot.New(
		&bot.Config{
			SlackBotToken: cb.config.SlackBotToken,
			SlackAppToken: cb.config.SlackAppToken,
			DB:            conn.DB,
			IDP:           idpclient,
			MDM:           mdmclient,
			Log:           cb.log,
			StatusChan:    cb.statusChan,
			StatusHandler: cb.statusHandler,
			Tables:        tables,
			LifeCycle:     cb,
			Method:        method,
			Cfg: bot.CfgSetter(
				bot.WithAuthUsers(cb.authUsers),
				bot.WithAuthUsersFromIDP(cb.flags.authUsersFromIDP),
				bot.WithCheckInterval(cb.flags.checkInterval),
				bot.WithClearTables(cb.flags.clearTables),
				bot.WithCutoffTime(cb.flags.cutoffTime),
				bot.WithDeadline(cb.flags.deadline),
				bot.WithDeviceDiffInterval(cb.flags.deviceDiffInterval),
				bot.WithHelpDocsURL(cb.flags.helpDocsURL),
				bot.WithHelpRepoURL(cb.flags.helpRepoURL),
				bot.WithHelpTicketURL(cb.flags.helpTicketURL),
				bot.WithLogLevel(cb.flags.logLevel),
				bot.WithLogToFile(cb.flags.logToFile),
				bot.WithPollInterval(cb.flags.pollInterval),
				bot.WithRequiredVers(cb.flags.requiredVers),
				bot.WithSlackAlertChannel(cb.config.SlackAlertChannel),
				bot.WithSlackBotID(cb.config.SlackBotID),
				bot.WithTableNames(cb.flags.tableNames),
				bot.WithTesting(cb.flags.testing),
				bot.WithTestingEndTime(cb.flags.testingEndTime),
				bot.WithTestingStartTime(cb.flags.testingStartTime),
				bot.WithTestUsers(cb.testUsers),
			),
		},
	)

	if cb.flags.authUsersFromIDP {
		oid, err := cb.idp.GetAdminGroup(cb.config.AdminGroupID)
		if err != nil {
			cb.log.Err(err).Msg("could not get admin group")
			os.Exit(3)
		}
		for _, e := range oid {
			sid, err := cb.bot.Client().GetUserByEmail(e)
			if err != nil {
				cb.log.Debug().Str("user", e).Msg("could not get user id")
				continue
			}
			cb.authUsers = append(cb.authUsers, sid.ID)
		}
	} else if cb.flags.authUsers != "" {
		userSlice := strings.Split(cb.flags.authUsers, ",")
		cb.authUsers = userSlice
	}

	cb.bot.UpdateCfg(bot.WithAuthUsers(cb.authUsers))

	return cb
}
