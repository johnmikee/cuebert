package timebound

import (
	"fmt"
	"math"
	"time"

	bi "github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

// Check implements method.Actions.
func (t *TimeBound) Check(time.Time) {
	t.check()
}

func (t *TimeBound) check() {
	devices, err := t.tables.GetBotTableInfo()
	if err != nil {
		return
	}

	for i := range devices {
		if !devices[i].FirstMessageSent {
			if devices[i].FirstMessageWaiting {
				t.log.Debug().
					Str("user", devices[i].UserEmail).
					Str("serial", devices[i].SerialNumber).
					Msg("skipping: routine for first message already started")

				continue
			}
			t.bot.SendReminder(1, &devices[i])
			continue
		}

		if !devices[i].FirstACK {
			// if the first message was sent but never ack'd
			// send it again.
			//
			// check to see if first message was just delivered
			// dont want to yell again if we are below the default reminder interval
			if !devices[i].FirstMessageSentAt.IsZero() {
				now := time.Now()
				diff := now.Sub(devices[i].FirstMessageSentAt)

				if diff.Minutes() < float64(t.cfg.defaultReminderInterval) {
					t.log.Trace().
						Str("user", devices[i].FullName).
						Time("first_message_sent_at", devices[i].FirstMessageSentAt).
						Float64("diff_hours", diff.Minutes()).
						Int("default_reminder_interval", t.cfg.defaultReminderInterval).
						Msg("skipping resend")
					continue
				}
			}
			t.bot.SendReminder(2, &devices[i])
		}

		_, err := t.checkReminders(devices[i])
		if err != nil {
			t.log.Err(err).Msg("could not check reminders")
			continue
		}

	}
	go t.statusHandler.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:          "check",
				Finish:        time.Now().Format(time.RFC3339),
				FinishNoError: true,
			},
			Start:  false,
			Finish: true,
			Err:    false,
		},
		"check",
	)
}

func (t *TimeBound) checkReminders(device bot.Info) (bool, error) {
	// the user has received the first message so
	// check how long its been since the first ack
	fa, err := t.tables.GetACKTime(device.SerialNumber)
	if err != nil {
		t.log.Err(err).Msg("could not get ack time")
		return false, nil
	}

	ack := fa.In(helpers.GenLocation(device.TZOffset))
	now := time.Now().In(helpers.GenLocation(device.TZOffset))

	diff := now.Sub(ack)

	t.log.Trace().
		Str("user", device.FullName).
		Str("serial", device.SerialNumber).
		Str("slack_id", device.SlackID).
		Int64("tz_offset", device.TZOffset).
		Float64("time_diff", diff.Hours()).
		Msg("time difference hours")

	if diff.Minutes() <= float64(t.cfg.defaultReminderInterval) {
		if !t.cfg.testing {
			return false, nil
		}
		if !helpers.Contains(t.cfg.testingUsers, device.SlackID) {
			return false, nil
		}
	} else {
		// if the time difference is greater than the default reminder interval
		i, err := t.tables.GetReminderInterval(device.SerialNumber)
		if err != nil {
			return false, fmt.Errorf("could not get reminder interval: %w", err)
		}

		dev, err := t.tables.DeviceBySerial(device.SerialNumber)
		if err != nil {
			// if we cant get the device info then we cant send the message
			t.log.Info().
				Str("slack_id", device.SlackID).
				Str("serial", device.SerialNumber).
				Str("user", device.FullName).
				AnErr("could not get devices", err).
				Send()
			return false, fmt.Errorf("could not get device info: %w", err)
		}

		// check if the user has a reminder waiting
		rw, err := t.tables.GetReminderWaiting(device.SerialNumber)
		if err != nil {
			return false, fmt.Errorf("could not get reminder waiting: %w", err)
		}

		if rw {
			return true, nil
		}

		distance := math.Abs(float64(diff.Minutes() - float64(i)))

		// we are close enough - spin off a routine to send the reminder
		if distance < 15 {
			go t.bot.ScheduleReminder(
				time.Duration(distance*float64(time.Minute)),
				&bi.ReminderInfo{
					Deadline: t.cfg.deadline,
					User:     device.SlackID,
					Serial:   device.SerialNumber,
					Version:  t.cfg.requiredVers,
					OS:       dev[0].OSVersion,
					Text:     t.reminderMessage(t.cfg.deadline),
				},
			)
			return true, nil
		}

		if diff.Minutes() >= float64(i) {
			t.ReminderMessage(
				&bi.ReminderPayload{
					UserSlackID: device.SlackID,
					UserName:    device.FullName,
					Serial:      device.SerialNumber,
					Model:       dev[0].Model,
					OS:          dev[0].OSVersion,
				},
			)
		}
	}
	return true, nil
}
