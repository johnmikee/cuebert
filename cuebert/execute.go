package main

import (
	"fmt"
	"time"
)

func (c *Cuebert) run() {
	c.log.Info().Msg("starting run")
	c.isRunning = true
	stop := make(chan struct{})

	c.logFlags()

	status := c.statusHandler.GetStatus()
	status.Message = "starting " + c.flags.serviceName
	status.Code = 200

	c.statusHandler.SetStatus(status)

	var check []string
	if c.flags.clearTables {
		err := c.tables.DeleteTables(c.flags.tableNames)
		if err != nil {
			c.log.Err(err).Msg("could not delete all tables")
		}

		check, err = c.tables.InitTables(c.flags.requiredVers)
		if err != nil {
			c.log.Err(err).Msg("could not initialize tables")
			c.stop()
			status := c.statusHandler.GetStatus()
			status.Message = fmt.Sprintf("stopping %s. could not build tables", c.flags.serviceName)
			status.Code = 400

			c.statusHandler.SetStatus(status)
		}

		go c.method.TableAssociations(check)
	}

	// check if we are past the deadline
	go c.doEvery(
		time.Duration(5)*time.Minute,
		c.checkDeadline,
	)
	// check periodically for changes on the devices.
	go c.doEvery(
		time.Duration(c.flags.deviceDiffInterval)*time.Minute,
		c.deviceDiff,
	)
	// here we check who needs the first reminder as well as the second message to the manager.
	go c.doEvery(
		time.Duration(c.flags.checkInterval)*time.Minute,
		c.method.Check,
	)
	// check if anyone who elected for a reminder needs a reminder
	go c.doEvery(
		time.Duration(c.flags.pollInterval)*time.Minute,
		c.method.Poll,
	)

	// Wait for stop signal
	<-c.stopSignal
	// Signal stop to the doEvery goroutines
	close(stop)
}

func (c *Cuebert) reloadFlags() {
	c.log.Info().Msg("reloading flags")
	c.logFlags()
}

func (c *Cuebert) start() {
	c.log.Trace().Msg("Starting bot...")
	c.startSignal <- struct{}{}
}

func (c *Cuebert) stop() {
	c.log.Trace().Msg("Stopping bot...")
	c.stopSignal <- struct{}{}
	c.isRunning = false
}

func (c *Cuebert) update() {
	c.log.Trace().Msg("Updating bot...")
	c.reloadSignal <- struct{}{}
}

// doEvery runs a function every on a duration until stop signal is received
func (c *Cuebert) doEvery(d time.Duration, f func(time.Time)) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for {
		select {
		case tm := <-ticker.C:
			if !c.isRunning {
				return
			}

			f(tm)

		case <-c.stopSignal:
			return
		}
	}
}
