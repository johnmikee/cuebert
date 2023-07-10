package main

import (
	"time"

	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

func (c *Cuebert) Poll(t time.Time) {
	br, err := c.tables.GetBotTableInfo()
	if err != nil {
		c.log.Err(err).Msg("could not get info from the bot results table")
		return
	}

	c.log.Debug().Msg("starting poll reminder")
	for i := range br {
		if !br[i].DelayAt.IsZero() {
			c.log.Debug().
				Str("user", br[i].SlackID).
				Str("serial", br[i].SerialNumber).
				Msg("has a reminder set - checking now")

			diff, err := helpers.LocaleDiff(br[i].TZOffset, br[i].DelayDate, br[i].DelayTime)
			if err != nil {
				c.log.Debug().
					Str("user", br[i].SlackID).
					Str("serial", br[i].SerialNumber).
					Int64("tz_offset", br[i].TZOffset).
					Str("delay_date", br[i].DelayDate).
					Str("delay_time", br[i].DelayTime).
					AnErr("could not get locale difference", err).
					Send()
				return
			}
			c.log.Debug().
				Str("user", br[i].SlackID).
				Str("serial", br[i].SerialNumber).
				Str("delay_at", br[i].DelayAt.String()).
				Str("delay_date", br[i].DelayDate).
				Float64("remind_in_minutes", diff.Abs().Minutes()).
				Msg("reminder set")

			// since this runs every 15 minutes or so if we are under that
			// spin off a routine to deal with it instead of risking missing it.
			//
			// if this is negative we dropped the ball.
			if diff.Abs().Minutes() <= float64(c.flags.pollInterval) {
				// make sure the delay hasnt already been sent
				if br[i].DelaySent {
					c.log.Debug().Str("user", br[i].SlackID).Msg("delay has already been sent")
					continue
				}

				c.log.Info().
					Str("user", br[i].SlackID).
					Str("serial", br[i].SerialNumber).
					Str("delay_at", br[i].DelayAt.String()).
					Str("delay_date", br[i].DelayDate).
					Float64("remind_in_minutes", diff.Abs().Minutes()).
					Msg("reminder set")

				// grab their device info
				var os string
				di, err := c.tables.DeviceBySerial(br[i].SerialNumber)
				if err != nil {
					c.log.Debug().
						Str("serial_number", br[i].SerialNumber).
						Str("user", br[i].SlackID).
						AnErr("could not get device", err).
						Send()

					os = "unknown"
				} else {
					os = di[0].OSVersion
				}
				// sleep until its time and then fire off the alert
				go c.bot.ScheduleReminder(diff,
					&bot.ReminderInfo{
						Deadline: c.flags.deadline,
						Cutoff:   c.flags.cutoffTime,
						User:     br[i].SlackID,
						Serial:   br[i].SerialNumber,
						Version:  c.flags.requiredVers,
						OS:       os,
						Text:     ":wave: Here is your requested reminder to update your device!",
					},
				)
			}
		}
	}

	// run method specific implementations
	c.method.Poll(t)

	go c.statusHandler.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:          "poll",
				Finish:        time.Now().Format(time.RFC3339),
				Message:       "finished poll reminder",
				FinishNoError: true,
			},
			Start:  false,
			Finish: true,
			Err:    false,
		},
		"poll",
	)
}
