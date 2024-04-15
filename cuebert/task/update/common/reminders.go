package common

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/devices"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/slack-go/slack"
)

var (
	TitleText  = slack.NewTextBlockObject(slack.PlainTextType, "Cuebert", false, false)
	CloseText  = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	SubmitText = slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
)

// ScheduleReminder will execute the scheduled reminder set by the user.
func (c *Common) scheduleReminder(t time.Duration, bi *tracking.Info) {
	if bi.ReminderWaiting {
		c.log.Trace().
			Str("user", bi.UserFullName).
			Bool("sent", bi.ReminderWaiting).
			Msg("not sending reminder, already sent")
		return
	}

	c.log.Trace().
		Str("user", bi.UserFullName).
		Float64("sleeping_seconds", t.Seconds()).
		Msg("waiting to send reminder")

	time.Sleep(t)

	c.log.Trace().
		Str("user", bi.UserFullName).
		Str("serial", bi.SerialNumber).
		Msg("sending requested reminder")

	err := c.shared.DeliverReminder(
		&shared.Reminder{
			Text:     "Cuebert Update Reminder",
			Deadline: c.flags.Deadline,
			Cutoff:   c.flags.CutoffTime,
			User:     bi.UserFullName,
			Task:     c.flags.RequiredVers + " Update",
			Title:    "Cuebert Update Reminder",
			Fields: []slack.AttachmentField{
				{
					Title: "Required Version",
					Value: c.flags.RequiredVers,
				},
				{
					Title: "Update Deadline",
					Value: c.flags.Deadline + " " + c.flags.CutoffTime,
				},
				{
					Title: "Serial Number",
					Value: bi.SerialNumber,
				},
			},
		},
	)
	if err != nil {
		c.log.Info().
			Str("user", bi.UserFullName).
			AnErr("delivering reminder", err).
			Send()
		return
	}

	err = c.tables.ReminderSent(true, bi.SerialNumber)
	if err != nil {
		c.log.Info().
			Str("user", bi.UserFullName).
			Str("serial", bi.SerialNumber).
			Str("table", "bot_results").
			AnErr("updating table", err).
			Send()
		return
	}

	c.log.Trace().
		Str("user", bi.UserFullName).
		Str("serial", bi.SerialNumber).
		Str("table", "bot_results").
		Bool("updated", true).
		Msg("notification reminder sent")
}

// SendReminder sends a reminder to the user based on their input
func (c *Common) SendReminder(count int, b *tracking.Info) {
	dev, err := c.tables.Device().By().Opts(devices.SerialNumber, b.SerialNumber).Query()
	if err != nil {
		c.log.Info().
			AnErr("getting devices", err).
			Str("user", b.UserSlackID).
			Send()
		return
	}

	if dev.Empty() {
		c.log.Info().
			Str("user", b.UserSlackID).
			Str("serial", b.SerialNumber).
			Msg("no device found")
		return
	}

	c.shared.SendMSG(
		&shared.ReminderPayload{
			UserSlackID: b.UserSlackID,
			UserName:    b.UserFullName,
			Item:        b.SerialNumber,
			TZOffset:    b.TZOffset,
		},
		count,
	)
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
