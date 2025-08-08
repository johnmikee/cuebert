package main

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func (c *Cuebert) run() {
	c.log.Info().Msg("starting run")

	// directly set the running state using atomic operations
	atomic.StoreInt32(&c.running, 1)
	c.lifecyle.Start()

	c.task.LogFlags()

	status := c.statusHandler.GetStatus()
	status.Message = "starting " + c.flags.ServiceName
	status.Code = 200

	c.statusHandler.SetStatus(status)

	// run the check once before starting the ticker
	c.task.Check(time.Now())

	jobs := c.task.Get()
	for _, job := range jobs {
		c.log.Trace().Str("name", job.Name).Interface("when", job.When).Msg("scheduling")
		go c.doEvery(job.When, job.Job, job.Continuous)
	}

	// wait for stop signal to perform cleanup and stop the application
	<-c.lifecyle.StopSignal()
	// it may be worth adding additional cleanup checks before stopping completely.
	// since the doEvery goroutines check the atomic running state, they will exit on their own when it's set to 0.
	atomic.StoreInt32(&c.running, 0) // ensure this happens as part of the lifecycle's Stop method or here directly
	c.lifecyle.Stop()
}

// doEvery runs a function every on a duration until stop signal is received.
func (c *Cuebert) doEvery(d time.Duration, f func(time.Time), continuous bool) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case tm := <-ticker.C:
			if atomic.LoadInt32(&c.running) == 0 {
				return
			}
			// check if the deadline has passed and send the stop signal if so
			if c.flags.Deadline != "" {
				deadlineMonthDay := strings.Split(c.flags.Deadline, "/")
				if len(deadlineMonthDay) == 2 {
					month, _ := strconv.Atoi(deadlineMonthDay[0])
					day, _ := strconv.Atoi(deadlineMonthDay[1])

					currentYear := time.Now().Year()
					deadlineTime := time.Date(currentYear, time.Month(month), day, 0, 0, 0, 0, time.Local)
					if time.Now().After(deadlineTime) {
						c.lifecyle.Stop()
						<-c.lifecyle.StopSignal()

						return
					}
				}
			}

			f(tm)

		case <-c.lifecyle.StopSignal():
			if !continuous {
				return
			}
		}
	}
}
