package main

import (
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
)

type reminderPayload struct {
	userSlackID    string
	userName       string
	managerSlackID string
	serial         string
	model          string
	os             string
	firstMessage   string
	tzOffset       int64
}

// check is the main function for the check routine. it checks the bot results
// table to see if and when users need to be reminded to update their devices.
// this handles the first message, acknowledgements, and second message with
// the user and their manager.
func (b *Bot) check(time.Time) {
	devices, err := b.db.GetBotTableInfo()
	if err != nil {
		return
	}

	for i := range devices {
		if !devices[i].FirstMessageSent {
			if devices[i].FirstMessageWaiting {
				b.log.Debug().
					Str("user", devices[i].UserEmail).
					Str("serial", devices[i].SerialNumber).
					Msg("skipping: routine for first message already started")

				continue
			}
			b.deliverReminder(1, &devices[i])
			continue
		}

		if !devices[i].FirstACK {
			// if the first message was sent but never ack'd
			// send it again.
			//
			// check to see if first message was just delivered
			// dont want to yell again if its only been a few hours
			if !devices[i].FirstMessageSentAt.IsZero() {
				now := time.Now()
				diff := now.Sub(devices[i].FirstMessageSentAt)

				if diff.Hours() < 24 {
					b.log.Trace().
						Str("user", devices[i].FullName).
						Time("first_message_sent_at", devices[i].FirstMessageSentAt).
						Float64("diff_hours", diff.Hours()).
						Msg("skipping resend")
					continue
				}
			}
			b.deliverReminder(2, &devices[i])
		}

		// the use has received the first message so
		// check how long its been since the first ack
		//
		// make sure the time isnt empty first though
		if devices[i].FirstACKTime.IsZero() {
			b.log.Trace().Msg("skipping zero time")
			continue
		}

		fa, err := b.db.GetACKTime(devices[i].SerialNumber)
		if err != nil {
			b.log.Err(err).Msg("could not get ack time")
			continue
		}

		ack := fa.In(genLocation(devices[i].TZOffset))
		now := time.Now().In(genLocation(devices[i].TZOffset))

		diff := now.Sub(ack)

		b.log.Trace().
			Str("user", devices[i].FullName).
			Str("serial", devices[i].SerialNumber).
			Str("slack_id", devices[i].SlackID).
			Int64("tz_offset", devices[i].TZOffset).
			Float64("time_diff", diff.Hours()).
			Msg("time difference hours")

		if now.Weekday() != time.Wednesday {
			// its been one week since you looked at me.. errr since the
			// first message was ack'd and its a wednesday.
			//
			// unless we are testing, then we just pick two days after the first ack.
			if !b.cfg.flags.testing {
				continue
			}
			if !helpers.Contains(b.cfg.testUsers, devices[i].SlackID) {
				continue
			}
		} else {
			if diff.Hours() >= 48 && !b.cfg.flags.testing && !helpers.Contains(b.cfg.testUsers, devices[i].SlackID) {
				continue
			}
			// first check if the manager has been notified
			s, err := b.db.GetManagerNotified(devices[i].SerialNumber)
			if err != nil {
				b.log.Err(err).Msg("could not get manager notified")
				continue
			}

			if s {
				// manager has already been notified
				b.log.Info().
					Str("user", devices[i].FullName).
					Str("serial", devices[i].SerialNumber).
					Str("manager_slack_id", devices[i].ManagerSlackID).
					Msg("manager already notified")

				return
			}

			dev, err := b.db.DeviceBySerial(devices[i].SerialNumber)
			if err != nil {
				// if we cant get the device info then we cant send the message
				b.log.Info().
					Str("slack_id", devices[i].SlackID).
					Str("serial", devices[i].SerialNumber).
					Str("user", devices[i].FullName).
					AnErr("could not get devices", err).
					Send()
				continue
			}

			// if we got here, the user has received the first message and ack'd it.
			fm := devices[i].FirstMessageSentAt.In(genLocation(devices[i].TZOffset))

			b.managerMessage(
				&reminderPayload{
					userSlackID:    devices[i].SlackID,
					userName:       devices[i].FullName,
					managerSlackID: devices[i].ManagerSlackID,
					serial:         devices[i].SerialNumber,
					model:          dev[0].Model,
					os:             dev[0].OSVersion,
					firstMessage:   fm.Format("Monday, January 2, 2006 3:04 PM"),
				},
			)
		}
	}
	go b.statusHandler.updateStatus(
		&routineUpdate{
			routine: &RoutineStatus{
				Name:          "check",
				Finish:        time.Now().Format(time.RFC3339),
				FinishNoError: true,
			},
			start:  false,
			finish: true,
			err:    false,
		},
		"check",
	)
}

// pollReminders checks the bot results table for any reminders that need to be sent.
// - these are the reminders set by the user.
func (b *Bot) pollReminders(time.Time) {
	br, err := b.db.GetBotTableInfo()
	if err != nil {
		b.log.Err(err).Msg("could not get info from the bot results table")
		return
	}

	b.log.Debug().Msg("starting poll reminder")
	for i := range br {
		if !br[i].DelayAt.IsZero() {
			b.log.Debug().
				Str("user", br[i].SlackID).
				Str("serial", br[i].SerialNumber).
				Msg("has a reminder set - checking now")

			diff, err := localeDiff(br[i].TZOffset, br[i].DelayDate, br[i].DelayTime)
			if err != nil {
				b.log.Debug().
					Str("user", br[i].SlackID).
					Str("serial", br[i].SerialNumber).
					Int64("tz_offset", br[i].TZOffset).
					Str("delay_date", br[i].DelayDate).
					Str("delay_time", br[i].DelayTime).
					AnErr("could not get locale difference", err).
					Send()
				return
			}
			b.log.Debug().
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
			if diff.Abs().Minutes() <= float64(b.cfg.flags.pollInterval) {
				// make sure the delay hasnt already been sent
				if br[i].DelaySent {
					b.log.Debug().Str("user", br[i].SlackID).Msg("delay has already been sent")
					continue
				}

				b.log.Info().
					Str("user", br[i].SlackID).
					Str("serial", br[i].SerialNumber).
					Str("delay_at", br[i].DelayAt.String()).
					Str("delay_date", br[i].DelayDate).
					Float64("remind_in_minutes", diff.Abs().Minutes()).
					Msg("reminder set")

				// grab their device info
				var os string
				di, err := b.db.DeviceBySerial(br[i].SerialNumber)
				if err != nil {
					b.log.Debug().
						Str("serial_number", br[i].SerialNumber).
						Str("user", br[i].SlackID).
						AnErr("could not get device", err).
						Send()

					os = "unknown"
				} else {
					os = di[0].OSVersion
				}
				// sleep until its t ime and then fire off the alert
				go b.scheduleReminder(diff,
					&reminderInfo{
						deadline: b.cfg.flags.deadline,
						cutoff:   b.cfg.flags.cutoffTime,
						user:     br[i].SlackID,
						serial:   br[i].SerialNumber,
						version:  b.cfg.flags.requiredVers,
						os:       os,
						text:     ":wave: Here is your requested reminder to update your device!",
						log:      b.log,
						bot:      b.bot,
					},
				)
			}
		}
	}
	go b.statusHandler.updateStatus(
		&routineUpdate{
			routine: &RoutineStatus{
				Name:          "poll",
				Finish:        time.Now().Format(time.RFC3339),
				Message:       "finished poll reminder",
				FinishNoError: true,
			},
			start:  false,
			finish: true,
			err:    false,
		},
		"poll",
	)
}
