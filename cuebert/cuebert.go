package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/db/create"
)

func main() {
	c := setup()
	// suppress slacker logs
	os.Stdout = nil

	c.log.Info().Msg("cuebert time!")

	c.log.Info().Msg("starting health handler...")
	go c.statusHandler.StartHealthHandler()

	if c.flags.dailyReport {
		err := c.bot.SendDailyAdminReport()
		status := c.statusHandler.GetStatus()
		if err != nil {
			c.log.Err(err).Msg("could not send daily report")
			status.DailyReport = &handlers.BotStatus{
				Message: "could not send daily report",
				Error:   err,
			}
		}
		status.DailyReport = &handlers.BotStatus{
			Message: "daily report sent",
			Error:   nil,
		}
		c.statusHandler.SetStatus(status)
		os.Exit(0)
	}

	// send cuebert off to handle questions
	go c.bot.Respond()
	// wait for a signal to start
	go c.handler()

	if !c.flags.init {
		c.start()
	} else {
		var check []string
		if c.flags.rebuildTablesOnFailure {
			c.log.Info().Msg("building tables")
			err := create.Build(c.db, &c.log)
			if err != nil {
				c.log.Err(err).Msg("could not build tables")
				os.Exit(3)
			}
			c.log.Info().Msg("initializing tables")
			check, err = c.tables.InitTables(c.flags.requiredVers)
			if err != nil {
				c.log.Err(err).Msg("could not initialize tables")
				os.Exit(3)
			}
		}
		go c.method.PostCheck(check)
	}

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	c.log.Info().Msg("cuebye!")
}

// control when we start and stop the main routines for the program
func (c *Cuebert) handler() {
	c.log.Info().Msg("starting routine handler...")
	// Wait for start or stop signal
	for {
		select {
		case <-c.reloadSignal:
			c.log.Info().Msg("updating config")
			go c.reloadFlags()
		case <-c.startSignal:
			// Start executing subsequent functions
			c.log.Info().Msg("starting routines")
			go c.run()
		case <-c.stopSignal:
			// Stop executing subsequent functions and wait for start signal again
			c.log.Info().Msg("stopping routines")
			c.isRunning = false
		}
	}
}
