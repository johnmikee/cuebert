package manager

import (
	"time"

	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

// check is the main function for the check routine. it checks the bot results
// table to see if and when users need to be reminded to update their devices.
// this handles the first message, acknowledgements, and second message with
// the user and their manager.
func (m *Manager) Check(time.Time) {
	devices, err := m.tables.GetBotTableInfo()
	if err != nil {
		return
	}

	for i := range devices {
		if !devices[i].FirstMessageSent {
			if devices[i].FirstMessageWaiting {
				m.log.Debug().
					Str("user", devices[i].UserEmail).
					Str("serial", devices[i].SerialNumber).
					Msg("skipping: routine for first message already started")

				continue
			}
			m.bot.SendReminder(1, &devices[i])
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
					m.log.Trace().
						Str("user", devices[i].FullName).
						Time("first_message_sent_at", devices[i].FirstMessageSentAt).
						Float64("diff_hours", diff.Hours()).
						Msg("skipping resend")
					continue
				}
			}
			m.bot.SendReminder(2, &devices[i])
		}

		// the use has received the first message so
		// check how long its been since the first ack
		//
		// make sure the time isnt empty first though
		if devices[i].FirstACKTime.IsZero() {
			m.log.Trace().Msg("skipping zero time")
			continue
		}

		fa, err := m.tables.GetACKTime(devices[i].SerialNumber)
		if err != nil {
			m.log.Err(err).Msg("could not get ack time")
			continue
		}

		ack := fa.In(helpers.GenLocation(devices[i].TZOffset))
		now := time.Now().In(helpers.GenLocation(devices[i].TZOffset))

		diff := now.Sub(ack)

		m.log.Trace().
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
			if !m.cfg.testing {
				continue
			}
			if !helpers.Contains(m.cfg.testingUsers, devices[i].SlackID) {
				continue
			}
		} else {
			if diff.Hours() >= 48 && !m.cfg.testing && !helpers.Contains(m.cfg.testingUsers, devices[i].SlackID) {
				continue
			}
			// first check if the manager has been notified
			s, err := m.tables.GetManagerNotified(devices[i].SerialNumber)
			if err != nil {
				m.log.Err(err).Msg("could not get manager notified")
				continue
			}

			if s {
				// manager has already been notified
				m.log.Info().
					Str("user", devices[i].FullName).
					Str("serial", devices[i].SerialNumber).
					Str("manager_slack_id", devices[i].ManagerSlackID).
					Msg("manager already notified")

				return
			}

			dev, err := m.tables.DeviceBySerial(devices[i].SerialNumber)
			if err != nil {
				// if we cant get the device info then we cant send the message
				m.log.Info().
					Str("slack_id", devices[i].SlackID).
					Str("serial", devices[i].SerialNumber).
					Str("user", devices[i].FullName).
					AnErr("could not get devices", err).
					Send()
				continue
			}

			// if we got here, the user has received the first message and ack'd it.
			fm := devices[i].FirstMessageSentAt.In(helpers.GenLocation(devices[i].TZOffset))

			m.managerMessage(
				&bot.ReminderPayload{
					UserSlackID:    devices[i].SlackID,
					UserName:       devices[i].FullName,
					ManagerSlackID: devices[i].ManagerSlackID,
					Serial:         devices[i].SerialNumber,
					Model:          dev[0].Model,
					OS:             dev[0].OSVersion,
					FirstMessage:   fm.Format("Monday, January 2, 2006 3:04 PM"),
				},
			)
		}
	}
	go m.statusHandler.UpdateStatus(
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

// DeviceDiff implements method.Actions.
func (m *Manager) DeviceDiff(sa []string) {
	m.log.Trace().Msg("not implemented")
}
