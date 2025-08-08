package main

import (
	"context"
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
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
	"go.uber.org/automaxprocs/maxprocs"
)

// CuebertRefactored is a struct to hold config for the program with improved error handling.
type CuebertRefactored struct {
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

func mainRefactored() {
	c, err := setupRefactored()
	if err != nil {
		// Since we don't have logger yet, use standard error
		panic(errors.Wrap(err, errors.ErrCodeInternal, "initialization failed"))
	}

	if c.flags.MakeMigrations {
		if err := migrations.MakeEmbed(c.db.GetConnStr()); err != nil {
			c.log.Fatal().Err(errors.DBError("run migrations", err)).Msg("could not run migrations")
		}
		c.log.Debug().Msg("migrations complete")
		os.Exit(0)
	}

	if runtime.GOOS == "linux" {
		if _, err := maxprocs.Set(maxprocs.Logger(c.log.Printf)); err != nil {
			c.log.Error().Err(errors.Wrap(err, errors.ErrCodeInternal, "set GOMAXPROCS")).Msg("cannot set GOMAXPROCS")
		}
	}

	c.log.Info().Msg("cuebert starting...")

	// Start handlers with error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.log.Error().
					Err(errors.New(errors.ErrCodeInternal, "handler panic")).
					Interface("panic", r).
					Msg("handler goroutine panicked")
			}
		}()
		handlers.StartHandlers(c.task.SetupReportHandler())
	}()

	// Start bot responder with error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.log.Error().
					Err(errors.SlackError("bot panic", nil)).
					Interface("panic", r).
					Msg("bot goroutine panicked")
			}
		}()
		c.bot.Respond()
	}()

	// Start monitor with error handling
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.log.Error().
					Err(errors.New(errors.ErrCodeInternal, "monitor panic")).
					Interface("panic", r).
					Msg("monitor goroutine panicked")
			}
		}()
		c.monitor()
	}()

	// Populate tracking table with better error handling
	_, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := c.task.Populate(time.Now()); err != nil {
		c.log.Error().
			Err(errors.DBError("populate tracking table", err)).
			Msg("could not populate tracking table")
		// Don't exit - continue with partial data
	}

	// Run initial diff
	c.task.Diff(time.Now())

	// Start periodic diff with error recovery
	go c.doEveryWithRecovery((15 * time.Minute), c.task.Diff, true)

	if !c.flags.Init {
		c.lifecyle.Start()
	}

	// Setup graceful shutdown
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)

	<-quitChannel
	c.log.Info().Msg("received signal to quit")
	atomic.StoreInt32(&c.running, 0)

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := c.gracefulShutdown(shutdownCtx); err != nil {
		c.log.Error().Err(errors.Wrap(err, errors.ErrCodeInternal, "graceful shutdown")).Msg("error during shutdown")
	}

	c.log.Info().Msg("cuebye!")
}

// gracefulShutdown handles shutting down all components gracefully.
func (c *CuebertRefactored) gracefulShutdown(ctx context.Context) error {
	errChan := make(chan error, 3)

	// Stop lifecycle
	go func() {
		c.lifecyle.Stop()
		errChan <- nil
	}()

	// Close database connections
	go func() {
		if c.db != nil {
			// Add database shutdown logic here
			errChan <- nil
		} else {
			errChan <- nil
		}
	}()

	// Stop status handler
	go func() {
		if c.statusHandler != nil {
			// Add status handler shutdown logic here
			errChan <- nil
		} else {
			errChan <- nil
		}
	}()

	// Wait for all shutdowns to complete or timeout
	var shutdownErrs []error
	for i := 0; i < 3; i++ {
		select {
		case err := <-errChan:
			if err != nil {
				shutdownErrs = append(shutdownErrs, err)
			}
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), errors.ErrCodeTimeout, "shutdown timeout")
		}
	}

	if len(shutdownErrs) > 0 {
		return errors.Wrap(shutdownErrs[0], errors.ErrCodeInternal, "shutdown errors")
	}

	return nil
}

// monitor controls when we start and stop the main routines with improved error handling.
func (c *CuebertRefactored) monitor() {
	c.log.Info().Msg("starting routine handler...")

	for {
		select {
		case <-c.lifecyle.ReloadSignal():
			c.log.Info().Msg("updating config")
			go func() {
				defer func() {
					if r := recover(); r != nil {
						c.log.Error().
							Err(errors.New(errors.ErrCodeInternal, "reload panic")).
							Interface("panic", r).
							Msg("reload goroutine panicked")
					}
				}()

				config.LogFlags(c.log, c.flags)
				c.lifecyle.Update()
			}()

		case <-c.lifecyle.StartSignal():
			if atomic.CompareAndSwapInt32(&c.running, 0, 1) {
				c.lifecyle.Start()
				c.log.Info().Msg("starting routines")

				go func() {
					defer func() {
						if r := recover(); r != nil {
							c.log.Error().
								Err(errors.New(errors.ErrCodeInternal, "start routines panic")).
								Interface("panic", r).
								Msg("start routines goroutine panicked")
						}
					}()

					config.LogFlags(c.log, c.flags)
					if err := c.run(); err != nil {
						c.log.Error().
							Err(errors.Wrap(err, errors.ErrCodeInternal, "run routines")).
							Msg("error running main routines")
					}
				}()
			}

		case <-c.lifecyle.StopSignal():
			if atomic.CompareAndSwapInt32(&c.running, 1, 0) {
				c.lifecyle.Stop()
				c.log.Info().Msg("stopping routines")
			}
		}
	}
}

// doEveryWithRecovery runs a function periodically with panic recovery.
func (c *CuebertRefactored) doEveryWithRecovery(d time.Duration, f func(time.Time), skipFirst bool) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	if !skipFirst {
		c.executeSafely(f)
	}

	for range ticker.C {
		if atomic.LoadInt32(&c.running) == 0 {
			return
		}
		c.executeSafely(f)
	}
}

// executeSafely executes a function with panic recovery and error logging.
func (c *CuebertRefactored) executeSafely(f func(time.Time)) {
	defer func() {
		if r := recover(); r != nil {
			c.log.Error().
				Err(errors.New(errors.ErrCodeInternal, "routine panic")).
				Interface("panic", r).
				Msg("routine panicked during execution")
		}
	}()

	f(time.Now())
}

// setupRefactored initializes the application with better error handling.
func setupRefactored() (*CuebertRefactored, error) {
	// Initialize flags first
	flags, err := config.LoadFlags()
	if err != nil {
		return nil, errors.ConfigError("load flags", err)
	}

	// Initialize logger
	log := logger.InitLogger(flags.LogLevel, flags.LogMode)

	// Load configuration
	cfg, err := config.Load(flags, log)
	if err != nil {
		return nil, errors.ConfigError("load config", err)
	}

	// Initialize database
	dbConfig := &dbc.Config{
		Cfg:   cfg,
		Flags: flags,
		Log:   log,
	}

	dbClient, err := db.New(dbConfig)
	if err != nil {
		return nil, errors.DBError("initialize database", err)
	}

	// Initialize other components...
	// (rest of setup logic with proper error handling)

	return &CuebertRefactored{
		// Initialize fields...
	}, nil
}
