package common

import (
	"context"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/lib/pq"
)

// Check pulls from the bot results table to see if and when users need to be reminded to update their devices.
func (c *Common) Check(t time.Time, p *shared.ReminderPayload) {
	tr, err := c.tables.Tracking().By().All().Query()
	if err != nil {
		return
	}

	if tr.Empty() { // ruh roh
		c.log.Debug().Msg("no users to check")
		return
	}

	for i := range tr {
		c.log.Trace().Interface("tracking", tr[i]).Msg("checking user")

		now := time.Now().In(helpers.GenLocation(tr[i].TZOffset))
		switch now.Weekday() {
		case time.Monday:
			// // check if the first message has been sent to the user already. certain user may have multiple devices
			// tableName := c.tables.Tracking().Table()

			// // Construct the SQL query string with the dynamic table name and use placeholders for other values.
			// query := fmt.Sprintf("SELECT serial_number, user_slack_id, user_email, first_message_sent FROM %s WHERE (user_email = $1 AND first_message_sent = true);", tableName)

			// // Execute the query with the parameters for the placeholders (except for the table name).
			// rows, err := c.tables.Pool().Query(context.Background(), query, tr[i].UserEmail)
			// if err != nil {
			// 	c.log.Debug().AnErr("error", err).Msg("could not get tracking info")
			// 	continue
			// }

			// type userTracking struct {
			// 	serialNumber     string
			// 	userSlackID      string
			// 	userEmail        string
			// 	firstMessageSent bool
			// }

			// var uts []userTracking
			// for rows.Next() {
			// 	var ut userTracking
			// 	err = rows.Scan(&ut.serialNumber, &ut.userSlackID, &ut.userEmail, &ut.firstMessageSent)
			// 	if err != nil {
			// 		c.log.Debug().AnErr("error", err).Msg("could not scan tracking info")
			// 		continue
			// 	}

			// 	uts = append(uts, ut)
			// }

			// if len(uts) > 0 { // if any first message has been sent for this user, skip
			// 	c.log.Debug().Str("user", tr[i].UserEmail).Msg("first message already sent for another device")
			// 	continue
			// }

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
				ti, err := c.tables.Tracking().By().Opts(tracking.UserSlackID, tr[i].UserSlackID).Query()
				if err != nil {
					c.log.Debug().AnErr("error", err).Msg("could not get tracking info")
					continue
				}

				if ti.Empty() {
					c.log.Debug().Msg("no tracking info found")
					continue
				}

				for x := range ti {
					if ti[x].FinalMessageSent {
						c.log.Debug().Str("serial", ti[x].SerialNumber).Msg("final message already sent for another device")
						continue
					}
				}

				time.Sleep(1 * time.Second)
				c.shared.SendMSG(&shared.ReminderPayload{
					UserServiceID: tr[i].UserServiceID,
					UserSlackID:   tr[i].UserSlackID,
					UserName:      tr[i].UserFullName,
					Task:          tr[i].TaskName,
					FinalMessage:  p.FinalMessage,
					Item:          p.Item,
					Text:          p.FinalMessage,
					TZOffset:      tr[i].TZOffset,
					Title:         p.Title,
				}, 3)
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

// Diff implements method.Actions.
func (c *Common) Diff(time.Time) {
	c.log.Trace().Msg("checking if we need to add any devices..")
	go c.status.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:    "Diff",
				Start:   time.Now().Format(time.RFC3339),
				Message: "starting diff routine",
			},
			Start:  true,
			Finish: false,
			Err:    false,
		}, "diff")

	md, err := c.mdm.ListDevices()
	if err != nil {
		c.log.Debug().AnErr("getting devices from mdm", err).Send()
		return
	}

	remove := []string{}
	for _, d := range md {
		ok, _ := helpers.CompareOSVer(d.OSVersion, c.flags.RequiredVers)
		if ok {
			c.log.Debug().Str("serial", d.SerialNumber).
				Str("os_vers", d.OSVersion).
				Msg("device is up to date - removing from tracking table")
			remove = append(remove, d.SerialNumber)
		} else {
			c.log.Debug().Str("serial", d.SerialNumber).
				Str("os_vers", d.OSVersion).
				Msg("device is not updated")
		}
	}

	// Log the serial numbers you intend to remove for diagnostic purposes
	c.log.Debug().Strs("serialsToRemove", remove).Msg("Preparing to remove devices from tracking table")

	// Check if there are any serial numbers to remove
	if len(remove) == 0 {
		c.log.Debug().Msg("No devices to remove from tracking table")
		return
	}

	// Use safe query builder to prevent SQL injection
	safeQueries := NewSafeQueries()
	tableName := c.tables.Tracking().Table()
	rowsAffected, err := safeQueries.DeleteTrackingRecords(context.Background(), c.tables.Pool(), tableName, remove)
	if err != nil {
		c.log.Debug().AnErr("removing devices from tracking table", err).Send()
		return
	}

	c.log.Debug().Int64("rowsAffected", rowsAffected).Msg("Removed devices from tracking table")

	go c.status.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:          "Diff",
				Finish:        time.Now().Format(time.RFC3339),
				Message:       "finished diff routine",
				FinishNoError: true,
			},
			Start:  false,
			Finish: true,
			Err:    false,
		},
		"diff",
	)
}

func (c *Common) Poll(t time.Time) {
	br, err := c.tables.Tracking().By().All().Query()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not get info from the bot results table")
		return
	}

	c.log.Debug().Msg("starting poll reminder...")
	for i := range br {
		if !br[i].UserReminderAt.IsZero() {
			c.log.Debug().
				Str("user", br[i].UserSlackID).
				Str("serial", br[i].SerialNumber).
				Msg("has a reminder set - checking now")

			diff, err := helpers.LocaleDiff(br[i].TZOffset, br[i].UserReminderDate, br[i].UserReminderTime)
			if err != nil {
				c.log.Debug().
					Str("user", br[i].UserSlackID).
					Int64("tz_offset", br[i].TZOffset).
					Str("delay_date", br[i].UserReminderDate).
					Str("delay_time", br[i].UserReminderTime).
					AnErr("could not get locale difference", err).
					Send()
				return
			}
			c.log.Debug().
				Str("user", br[i].UserSlackID).
				Str("delay_at", br[i].UserReminderAt.String()).
				Str("delay_date", br[i].UserReminderDate).
				Float64("remind_in_minutes", diff.Abs().Minutes()).
				Msg("reminder set")

			// since this runs every 15 minutes or so if we are under that
			// spin off a routine to deal with it instead of risking missing it.
			//
			// if this is negative we dropped the ball.
			if diff.Abs().Minutes() <= float64(c.flags.PollInterval) {
				// make sure the delay hasnt already been sent
				if br[i].ReminderSent {
					c.log.Debug().Str("user", br[i].UserSlackID).Msg("delay has already been sent")
					continue
				}

				c.log.Info().
					Str("user", br[i].UserSlackID).
					Str("delay_at", br[i].UserReminderAt.String()).
					Str("delay_date", br[i].UserReminderDate).
					Float64("remind_in_minutes", diff.Abs().Minutes()).
					Msg("reminder set")

				go c.scheduleReminder(diff, &br[i])
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
