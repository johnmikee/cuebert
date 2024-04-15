package common

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/slack-go/slack"
)

// SendReminder sends a reminder to the user based on their input
func (c *Common) SendReminder(count int, t *tracking.Info) {
	c.shared.SendMSG(
		&shared.ReminderPayload{
			UserSlackID:   t.UserSlackID,
			UserName:      t.UserFullName,
			UserServiceID: t.UserServiceID,
			TZOffset:      t.TZOffset,
		},
		count,
	)
}

// ScheduleReminder will execute the scheduled reminder set by the user.
func (c *Common) ScheduleReminder(t time.Duration, ti *tracking.Info) {
	if ti.ReminderWaiting {
		c.log.Debug().
			Str("user", ti.UserFullName).
			Bool("sent", ti.ReminderWaiting).
			Msg("not sending reminder, already sent")
		return
	}

	c.log.Trace().
		Str("user", ti.UserFullName).
		Float64("sleeping_seconds", t.Seconds()).
		Msg("waiting to send reminder")

	time.Sleep(t)

	c.log.Debug().
		Str("user", ti.UserFullName).
		Msg("sending requested reminder")

	err := c.shared.DeliverReminder(
		&shared.Reminder{
			Text:     "Cuebert Update Reminder",
			Deadline: c.flags.Deadline,
			Cutoff:   c.flags.CutoffTime,
			User:     ti.UserSlackID,
			Task:     ti.TaskName,
			Title:    "Hi there! Just a gentle reminder to acknowledge the previous message about finishing the task.",
			Fields:   []slack.AttachmentField{},
		},
	)
	if err != nil {
		c.log.Info().
			Str("user", ti.UserFullName).
			AnErr("delivering reminder", err).
			Send()
		return
	}

	_, err = c.tables.Tracking().Update().WithOptions(
		tracking.Info{UserSlackID: ti.UserSlackID, ReminderSent: true}).Send()
	if err != nil {
		c.log.Info().
			Str("user", ti.UserFullName).
			AnErr("updating table", err).
			Send()
		return
	}

	c.log.Trace().
		Str("user", ti.UserFullName).
		Bool("updated", true).
		Msg("notification reminder sent")
}

// UpdateReminderTime updates the reminder time in the db.
func (c *Common) UpdateReminderTime(dateVal, timeVal, slackID string) error {
	dateString := fmt.Sprintf("%s %s", dateVal, timeVal)
	ts, err := time.Parse("2006-01-02 15:04", dateString)
	if err != nil {
		c.log.Err(err).Msg("could not compose time string")
		return err
	}

	return c.tables.SetReminder(slackID, ts.UTC(), dateVal, timeVal)
}
