package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/johnmikee/cuebert/db/create"
	"github.com/johnmikee/cuebert/idp"
	idpclient "github.com/johnmikee/cuebert/idp/client"
	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/internal/device"
	"github.com/johnmikee/cuebert/internal/user"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/mdm/client"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker"
)

func main() {
	cfg, _ := loadEnv()
	// suppress slacker logs
	os.Stdout = nil

	cfg.log.Info().Msg("cuebert time!")
	conn, err := db.Connect(
		&db.Conf{
			Host:     cfg.DBAddress,
			Name:     cfg.DBName,
			Password: cfg.DBPass,
			Port:     cfg.DBPort,
			User:     cfg.DBUser,
		},
	)
	if err != nil {
		cfg.log.Info().AnErr("connecting to db", err).Send()
		os.Exit(3)
	}

	mdmclient := client.New(
		&client.MDM{
			MDM: mdm.MDM(cfg.flags.mdm),
			Config: mdm.Config{
				Domain:                 cfg.IDPDomain,
				MDM:                    mdm.MDM(cfg.flags.mdm),
				URL:                    cfg.MDMURL,
				Token:                  cfg.MDMKey,
				Client:                 nil,
				Log:                    cfg.log,
				ProviderSpecificConfig: nil,
			},
		},
	)
	// set the bot up
	bot := &Bot{
		bot:      slacker.NewClient(cfg.SlackBotToken, cfg.SlackAppToken, slacker.WithDebug(false)),
		cfg:      cfg,
		db:       db.New(conn.DB, &cfg.log),
		commands: []Commands{},
		log:      logger.ChildLogger("cuebert_bot", &cfg.log),
		idp: idpclient.New(
			&idpclient.IDP{
				IDP: idp.IDP(cfg.flags.idp),
				Config: idp.Config{
					Domain: cfg.IDPDomain,
					URL:    cfg.IDPURL,
					Token:  cfg.IDPToken,
					Client: nil,
					Log:    cfg.log,
				},
			},
		),
		mdm: mdmclient,
		tables: &Tables{
			db: db.New(conn.DB, &cfg.log),
			devices: device.New(
				&device.Config{
					DB:     conn.DB,
					Client: mdmclient,
					Log:    &cfg.log,
				},
			),
			log: logger.ChildLogger("tables", &cfg.log),
			users: user.New(
				&user.UserConfig{
					Client:     mdmclient,
					SlackToken: cfg.SlackBotToken,
					SlackUrl:   "https://slack.com/api/",
					DB:         conn.DB,
					Log:        &cfg.log,
				},
			),
		},
		startSignal:   make(chan struct{}),
		reloadSignal:  make(chan struct{}),
		stopSignal:    make(chan struct{}),
		statusHandler: &StatusHandler{},
		statusChan:    make(chan StatusMessage),
	}

	if bot.cfg.flags.authUsersFromIDP {
		oid, err := bot.idp.GetAdminGroup(cfg.AdminGroupID)
		if err != nil {
			bot.log.Err(err).Msg("could not get admin group")
			os.Exit(3)
		}
		for _, e := range oid {
			sid, err := bot.bot.Client().GetUserByEmail(e)
			if err != nil {
				bot.log.Debug().Str("user", e).Msg("could not get user id")
				continue
			}
			bot.cfg.authUsers = append(bot.cfg.authUsers, sid.ID)
		}
	} else if bot.cfg.flags.authUsers != "" {
		userSlice := strings.Split(bot.cfg.flags.authUsers, ",")
		bot.cfg.authUsers = userSlice
	}

	bot.log.Info().Msg("starting health handler...")
	go bot.statusHandler.StartHealthHandler()

	if cfg.flags.dailyReport {
		err := bot.SendDailyAdminReport()
		status := bot.statusHandler.GetStatus()
		if err != nil {
			bot.log.Err(err).Msg("could not send daily report")
			status.DailyReport = &BotStatus{
				Message: "could not send daily report",
				Error:   err,
			}
		}
		status.DailyReport = &BotStatus{
			Message: "daily report sent",
			Error:   nil,
		}
		bot.statusHandler.SetStatus(status)
		os.Exit(0)
	}

	// send cuebert off to handle questions
	go bot.respond()
	// wait for a signal to start
	go bot.handler()

	if !cfg.flags.init {
		bot.start()
	} else {
		var check []string
		if bot.cfg.flags.rebuildTablesOnFailure {
			bot.log.Info().Msg("building tables")
			err := create.Build(conn.DB, &cfg.log)
			if err != nil {
				bot.log.Err(err).Msg("could not build tables")
				os.Exit(3)
			}
			bot.log.Info().Msg("initializing tables")
			check, err = bot.tables.initTables(cfg.flags.requiredVers)
			if err != nil {
				bot.log.Err(err).Msg("could not initialize tables")
				os.Exit(3)
			}
		}
		go associateUserManager(bot, check)
	}

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	bot.log.Info().Msg("cuebye!")
}

// control when we start and stop the main routines for the program
func (b *Bot) handler() {
	b.log.Info().Msg("starting routine handler...")
	// Wait for start or stop signal
	for {
		select {
		case <-b.reloadSignal:
			b.log.Info().Msg("updating config")
			go b.reloadFlags()
		case <-b.startSignal:
			// Start executing subsequent functions
			b.log.Info().Msg("starting routines")
			go b.run()
		case <-b.stopSignal:
			// Stop executing subsequent functions and wait for start signal again
			b.log.Info().Msg("stopping routines")
			b.isRunning = false
		}
	}
}

// run the bot
func (b *Bot) run() {
	b.log.Info().Msg("starting run")
	b.isRunning = true
	stop := make(chan struct{})

	b.logFlags()

	status := b.statusHandler.GetStatus()
	status.Message = "starting " + b.cfg.flags.serviceName
	status.Code = 200

	b.statusHandler.SetStatus(status)

	var check []string
	if b.cfg.flags.clearTables {
		err := b.tables.deleteTables(b.cfg.flags.tableNames)
		if err != nil {
			b.log.Err(err).Msg("could not delete all tables")
		}

		check, err = b.tables.initTables(b.cfg.flags.requiredVers)
		if err != nil {
			b.log.Err(err).Msg("could not initialize tables")
			b.stop()
			status := b.statusHandler.GetStatus()
			status.Message = fmt.Sprintf("stopping %s. could not build tables", b.cfg.flags.serviceName)
			status.Code = 400

			b.statusHandler.SetStatus(status)
		}
		// associate users and their managers
		go associateUserManager(b, check)
	}

	// check periodically for changes on the devices.
	go b.doEvery(
		time.Duration(b.cfg.flags.deviceDiffInterval)*time.Minute,
		b.deviceDiff,
	)
	// here we check who needs the first reminder as well as the second message to the manager.
	go b.doEvery(
		time.Duration(b.cfg.flags.checkInterval)*time.Minute,
		b.check,
	)
	// check if anyone who elected for a reminder needs a reminder
	go b.doEvery(
		time.Duration(b.cfg.flags.pollInterval)*time.Minute,
		b.pollReminders,
	)

	// Wait for stop signal
	<-b.stopSignal
	// Signal stop to the doEvery goroutines
	close(stop)
}

func (b *Bot) reloadFlags() {
	b.log.Info().Msg("reloading flags")
	b.logFlags()
}
func (b *Bot) start() {
	b.log.Trace().Msg("Starting bot...")
	b.startSignal <- struct{}{}
}

func (b *Bot) stop() {
	b.log.Trace().Msg("Stopping bot...")
	b.stopSignal <- struct{}{}
	b.isRunning = false
}

func (b *Bot) update() {
	b.log.Trace().Msg("Updating bot...")
	b.reloadSignal <- struct{}{}
}

// doEvery runs a function every on a duration until stop signal is received
func (b *Bot) doEvery(d time.Duration, f func(time.Time)) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case tm := <-ticker.C:
			if !b.isRunning {
				return
			}

			f(tm)

		case <-b.stopSignal:
			return
		}
	}
}
