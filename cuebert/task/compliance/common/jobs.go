package common

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/db/shared/exclusions"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

// CheckInfo holds the information needed to check if the message should be sent.
type CheckInfo struct {
	ts      tracking.Info
	now     time.Time
	message string
}

// Check is the main function for the check routine.
//
// it queries the tracking table to see if and when users need to be reminded to finish updating.
func (c *Common) Check(t time.Time, p *shared.ReminderPayload) {
	c.log.Info().Msg("starting check routine...")
	tr, err := c.tables.Tracking().By().All().Query()
	if err != nil {
		return
	}

	if tr.Empty() {
		c.log.Debug().Msg("no users to check")
		return
	}

	for i := range tr {
		c.log.Trace().Interface("tracking", tr[i]).Msg("checking")
		now := time.Now().In(helpers.GenLocation(tr[i].TZOffset))

		// make sure the user is not in the excluded list
		ex, err := c.tables.Exclusion().By().Opts(exclusions.UserEmail, tr[i].UserEmail).Query()
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not check exclusion table")
			continue
		}
		if !ex.Empty() {
			if ex[0].Approved {
				c.log.Info().Str("user", tr[i].UserEmail).Msg("user is excluded - skipping")
				continue
			} else {
				c.log.Info().Str("user", tr[i].UserEmail).Msg("user is excluded but not approved - sending message")
			}
		}

		switch now.Weekday() {
		case time.Monday:
			if !tr[i].FirstMessageSent {
				c.log.Debug().Interface("tracking", tr[i]).Msg("sending first message")
				if tr[i].FirstMessageWaiting {
					c.log.Debug().
						Str("user", tr[i].UserEmail).
						Bool("task_complete", tr[i].TaskComplete).
						Msg("skipping: routine for first message already started")

					continue
				}
				c.shared.SendMSG(
					&shared.ReminderPayload{
						UserServiceID: tr[i].UserServiceID,
						UserSlackID:   tr[i].UserSlackID,
						UserName:      tr[i].UserFullName,
						FirstMessage:  p.FirstMessage,
						TZOffset:      tr[i].TZOffset,
						Title:         p.Title,
					}, 1)
				continue

			}

		case time.Friday:
			if !tr[i].FinalMessageSent {
				c.finalMessage(
					&CheckInfo{
						ts:      tr[i],
						now:     now,
						message: p.FinalMessage,
					},
				)
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

// Diff keeps the tracking table populated only with the users that
// have not completed the task. All users are added to the task table which we
// use to track adoption rate through the week.
//
// First we update the task table and then remove any of the users who have
// completed the task from the tracking table.
func (c *Common) Diff(time.Time) {
	c.log.Info().Msg("starting diff routine...")
	ti, err := c.tables.Task().By().All().Query()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not get info from the tasks table")
		return
	}

	if ti.Empty() {
		c.log.Debug().Msg("no users to diff")
		return
	}

	remove := ti.Completed()

	tr, err := c.tables.Tracking().By().All().Query()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not get info from the tracking table")
		return
	}

	if tr.Empty() {
		c.log.Debug().Msg("no users in tracking to remove")
		return
	}

	var removeTracking []tracking.Info
	for i := range remove {
		for j := range tr {
			if remove[i].UserID == tr[j].UserServiceID {
				removeTracking = append(removeTracking, tr[j])
			}
		}
	}

	if len(removeTracking) == 0 {
		c.log.Debug().Msg("no users to remove from tracking")
		return
	}

	for i := range removeTracking {
		_, err := c.tables.Tracking().Remove().
			Where(
				tracking.UserServiceID,
				removeTracking[i].UserServiceID,
			).Execute()
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not remove user from tracking table")
			continue
		}
	}

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
		"check",
	)
}

// Poll checks the tracking table for users that have a custom reminder set.
func (c *Common) Poll(t time.Time) {
	c.log.Info().Msg("starting poll reminder...")
	tr, err := c.tables.Tracking().By().All().Query()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not get info from the tracking table")
		return
	}

	if tr.Empty() {
		c.log.Debug().Msg("no users to poll")
		return
	}

	for i := range tr {
		if !tr[i].DelayAt.IsZero() {
			c.log.Debug().
				Str("user", tr[i].UserSlackID).
				Msg("has a reminder set - checking now")

			diff, err := helpers.LocaleDiff(tr[i].TZOffset, tr[i].DelayDate, tr[i].DelayTime)
			if err != nil {
				c.log.Debug().
					Str("user", tr[i].UserSlackID).
					Int64("tz_offset", tr[i].TZOffset).
					Str("delay_date", tr[i].DelayDate).
					Str("delay_time", tr[i].DelayTime).
					AnErr("could not get locale difference", err).
					Send()
				return
			}
			c.log.Debug().
				Str("user", tr[i].UserSlackID).
				Str("delay_at", tr[i].DelayAt.String()).
				Str("delay_date", tr[i].DelayDate).
				Float64("remind_in_minutes", diff.Abs().Minutes()).
				Msg("reminder set")

			// if this is negative we dropped the ball.
			if diff.Abs().Minutes() <= float64(c.flags.PollInterval) {
				// make sure the delay hasnt already been sent
				if tr[i].ReminderSent {
					c.log.Debug().Str("user", tr[i].UserSlackID).Msg("delay has already been sent")
					continue
				}

				c.log.Info().
					Str("user", tr[i].UserSlackID).
					Str("delay_at", tr[i].DelayAt.String()).
					Str("delay_date", tr[i].DelayDate).
					Float64("remind_in_minutes", diff.Abs().Minutes()).
					Msg("reminder set")

				// sleep until its time and then fire off the alert
				go c.ScheduleReminder(diff, &tr[i])
			}
		}
	}

	go c.status.UpdateStatus(
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

func (c *Common) finalMessage(i *CheckInfo) {
	if c.flags.Testing {
		if !helpers.Contains(c.flags.TestUsers, i.ts.UserSlackID) {
			return
		}
	}

	sn, err := rand.Int(rand.Reader, big.NewInt(60))
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not generate random delay")
	}

	go func() {
		err := c.FinalMessage(
			&shared.ReminderPayload{
				UserServiceID: i.ts.UserServiceID,
				UserSlackID:   i.ts.UserSlackID,
				UserName:      i.ts.UserFullName,
				Text:          i.message,
				TZOffset:      i.ts.TZOffset,
				Title:         "Final Reminder",
			}, sn.Int64())
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not send final message")
		}
	}()
}
