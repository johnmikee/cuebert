package common

import (
	"context"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/db/repository"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

// CheckRefactored pulls from the bot results table to see if and when users need to be reminded to update their devices.
// This is the refactored version using the repository pattern.
func (c *Common) CheckRefactored(t time.Time, p *shared.ReminderPayload) {
	ctx := context.Background()

	// Create tracking repository instance
	trackingRepo := repository.NewTrackingRepository(c.tables.Pool(), c.log, repository.TrackingTypeUpdate)

	// Get all tracking records
	trackingRecords, err := trackingRepo.FindAll(ctx, nil)
	if err != nil {
		c.log.Error().AnErr("error", errors.DBError("fetch tracking records", err)).Msg("failed to get tracking records")
		return
	}

	if len(trackingRecords) == 0 {
		c.log.Debug().Msg("no users to check")
		return
	}

	for _, tr := range trackingRecords {
		c.log.Trace().Interface("tracking", tr).Msg("checking user")

		now := time.Now().In(helpers.GenLocation(tr.TZOffset))
		switch now.Weekday() {
		case time.Monday:
			if !tr.FirstMessageSent {
				c.log.Debug().Interface("tracking", tr).Msg("sending first message")
				if tr.FirstMessageWaiting {
					c.log.Debug().
						Str("user", tr.UserEmail).
						Bool("task_complete", tr.TaskComplete).
						Msg("skipping: routine for first message already started")
					continue
				}

				// Send first message
				c.shared.SendMSG(
					&shared.ReminderPayload{
						UserServiceID: tr.UserServiceID,
						UserSlackID:   tr.UserSlackID,
						UserName:      tr.UserFullName,
						FirstMessage:  p.FirstMessage,
						TZOffset:      tr.TZOffset,
						Title:         p.Title,
					}, 1)
				continue
			}

		case time.Friday:
			if !tr.FinalMessageSent {
				// Check if final message already sent for any device for this user
				userTrackingRecords, err := trackingRepo.FindByUserSlackID(ctx, tr.UserSlackID)
				if err != nil {
					c.log.Debug().AnErr("error", errors.DBError("get user tracking", err)).Msg("could not get tracking info")
					continue
				}

				if len(userTrackingRecords) == 0 {
					c.log.Debug().Msg("no tracking info found")
					continue
				}

				// Check if any device already has final message sent
				finalMessageAlreadySent := false
				for _, record := range userTrackingRecords {
					if record.FinalMessageSent {
						c.log.Debug().Str("serial", record.SerialNumber).Msg("final message already sent for another device")
						finalMessageAlreadySent = true
						break
					}
				}

				if !finalMessageAlreadySent {
					time.Sleep(1 * time.Second)
					c.shared.SendMSG(&shared.ReminderPayload{
						UserServiceID: tr.UserServiceID,
						UserSlackID:   tr.UserSlackID,
						UserName:      tr.UserFullName,
						Task:          tr.TaskName,
						FinalMessage:  p.FinalMessage,
						Item:          p.Item,
						Text:          p.FinalMessage,
						TZOffset:      tr.TZOffset,
						Title:         p.Title,
					}, 3)
				}
			}

		default:
			c.log.Info().Msg("no messages to send")
		}
	}

	go c.status.UpdateStatus(
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

// DiffRefactored implements method.Actions using the repository pattern.
func (c *Common) DiffRefactored(time.Time) {
	c.log.Trace().Msg("checking if we need to add any devices..")
	go c.status.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:  "diff",
				Start: time.Now().Format(time.RFC3339),
			},
			Start:  true,
			Finish: false,
			Err:    false,
		},
		"diff",
	)

	defer func() {
		go c.status.UpdateStatus(
			&handlers.RoutineUpdate{
				Routine: &handlers.RoutineStatus{
					Name:          "diff",
					Finish:        time.Now().Format(time.RFC3339),
					FinishNoError: true,
				},
				Start:  false,
				Finish: true,
				Err:    false,
			},
			"diff",
		)
	}()

	// The diff logic would continue here, using repository pattern for database operations
	// This is a placeholder for the actual implementation
	c.updateDevices()
}

// PollRefactored checks for pending reminders using the repository pattern.
func (c *Common) PollRefactored(t time.Time) {
	ctx := context.Background()
	trackingRepo := repository.NewTrackingRepository(c.tables.Pool(), c.log, repository.TrackingTypeUpdate)

	// Find users with pending reminders
	pendingReminders, err := trackingRepo.FindUsersWithPendingReminders(ctx)
	if err != nil {
		c.log.Error().AnErr("error", errors.DBError("find pending reminders", err)).Msg("failed to get pending reminders")
		return
	}

	go c.status.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:  "poll",
				Start: time.Now().Format(time.RFC3339),
			},
			Start:  true,
			Finish: false,
			Err:    false,
		},
		"poll",
	)

	for _, reminder := range pendingReminders {
		c.log.Debug().
			Str("user", reminder.UserEmail).
			Str("delay_date", reminder.DelayDate).
			Str("delay_time", reminder.DelayTime).
			Msg("checking reminder")

		// Check if it's time to send the reminder
		now := time.Now().In(helpers.GenLocation(reminder.TZOffset))
		delayDateTime, err := time.ParseInLocation("2006-01-02 15:04", reminder.DelayDate+" "+reminder.DelayTime, now.Location())
		if err != nil {
			c.log.Error().AnErr("error", err).Msg("failed to parse delay time")
			continue
		}

		if now.After(delayDateTime) {
			// Send the reminder
			c.shared.SendMSG(&shared.ReminderPayload{
				UserServiceID: reminder.UserServiceID,
				UserSlackID:   reminder.UserSlackID,
				UserName:      reminder.UserFullName,
				Task:          reminder.TaskName,
				ReminderText:  "Reminder: Please complete your task",
				TZOffset:      reminder.TZOffset,
			}, 2)

			// Update reminder status
			err = trackingRepo.UpdateReminderStatus(ctx, reminder.UserSlackID, reminder.TaskName, true)
			if err != nil {
				c.log.Error().AnErr("error", errors.DBError("update reminder status", err)).Msg("failed to update reminder status")
			}
		}
	}

	go c.status.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:          "poll",
				Finish:        time.Now().Format(time.RFC3339),
				FinishNoError: true,
			},
			Start:  false,
			Finish: true,
			Err:    false,
		},
		"poll",
	)
}
