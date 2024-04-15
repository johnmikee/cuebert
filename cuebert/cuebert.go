package main

import (
	"os"
	"os/signal"
	"runtime"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/bot/lifecycle"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/db"
	dbc "github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/db/migrations"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"go.uber.org/automaxprocs/maxprocs"
)

// Cuebert is a struct to hold config for the program
type Cuebert struct {
	bot           *bot.Bot
	db            db.Client
	dbConfig      *dbc.Config
	config        *config.Config
	flags         *config.Flags
	log           logger.Logger
	lifecyle      *lifecycle.Client
	task          task.Rabbit
	statusChan    chan handlers.StatusMessage
	statusHandler *handlers.StatusHandler
	running       int32 // used for atomic operations
}

func main() {
	c := setup()

	if c.flags.MakeMigrations {
		err := migrations.MakeEmbed(c.db.GetConnStr())
		if err != nil {
			c.log.Fatal().AnErr("error", err).Msg("could not run migrations")
		}
		c.log.Debug().Msg("migrations complete")
		os.Exit(0)
	}
	if runtime.GOOS == "linux" { // set GOMAXPROCS to match Linux container CPU quota on Linux.
		if _, err := maxprocs.Set(maxprocs.Logger(c.log.Printf)); err != nil {
			c.log.Error().Msg("cannot set GOMAXPROCS")
		}
	}
	c.log.Info().Msg("cuebert starting...")

	go handlers.StartHandlers(c.task.SetupReportHandler())

	// send cuebert off to handle questions
	go c.bot.Respond()
	go c.monitor() // wait for a signal to start

	err := c.task.Populate(time.Now())
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not populate tracking table")
	}
	c.task.Diff(time.Now())

	go c.doEvery((15 * time.Minute), c.task.Diff, true)

	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not populate tracking table")
	}

	if !c.flags.Init {
		c.lifecyle.Start()
	}
	quitChannel := make(chan os.Signal, 1) // sized channel to avoid blocking
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)

	<-quitChannel
	c.log.Info().Msg("received signal to quit")
	atomic.StoreInt32(&c.running, 0)
	c.lifecyle.Stop()
	c.log.Info().Msg("cuebye!")
}

// control when we start and stop the main routines for the program
func (c *Cuebert) monitor() {
	c.log.Info().Msg("starting routine handler...")
	// wait for start, stop, or reload signals from the lifecycle system
	for {
		select {
		case <-c.lifecyle.ReloadSignal():
			c.log.Info().Msg("updating config")
			go func() {
				config.LogFlags(c.log, c.flags)
				c.lifecyle.Update() // trigger an update in the lifecycle system.
			}()

		case <-c.lifecyle.StartSignal():
			if atomic.CompareAndSwapInt32(&c.running, 0, 1) { // ensure not already running
				c.lifecyle.Start()
				c.log.Info().Msg("starting routines")
				go config.LogFlags(c.log, c.flags)
				go c.run() // start the main application routine.
			}

		case <-c.lifecyle.StopSignal():
			if atomic.CompareAndSwapInt32(&c.running, 1, 0) { // ensure it's currently running
				c.lifecyle.Stop()
				c.log.Info().Msg("stopping routines")
				// we can add logic to wait for routines to clean up before completely stopping.
			}
		}
	}
}
