package common

import (
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/slack-go/slack"
)

// FirstMessageSent sets the first message sent flag for the given slack id.
func (c *Common) FirstMessageSent(serviceID, slackID string) error {
	_, err := c.tables.Tracking().Update().WithOptions(
		tracking.Info{
			UserServiceID:       serviceID,
			UserSlackID:         slackID,
			FirstMessageWaiting: true,
		}).Parse("user_service_id", serviceID).Send()

	return err
}

// FirstMessageDelivered sets the first message delivered flag for the given slack id.
func (c *Common) FirstMessageDelivered(serviceID, slackID string) error {
	_, err := c.tables.Tracking().Update().WithOptions(
		tracking.Info{
			UserServiceID:      serviceID,
			UserSlackID:        slackID,
			FirstMessageSent:   true,
			FirstMessageSentAt: helpers.UpdateTime(),
		}).Parse(tracking.UserSlackID, slackID).Send()

	return err
}

// FinalMessageSent sets the final message sent flag for the given slack id.
func (c *Common) FinalMessageSent(serviceID, slackID string) error {
	_, err := c.tables.Tracking().Update().WithOptions(
		tracking.Info{
			UserServiceID:      serviceID,
			UserSlackID:        slackID,
			FinalMessageSent:   true,
			FinalMessageSentAt: helpers.UpdateTime(),
		}).Parse(tracking.UserSlackID, slackID).Send()

	return err
}

// ReminderMessage delivers the reminder message.
func (c *Common) ReminderMessage(rp *shared.ReminderPayload) error {
	return c.shared.DeliverReminder(
		&shared.Reminder{
			Text:     "Hi there! Just a gentle reminder to acknowledge the previous message about updating.",
			Deadline: c.flags.Deadline,
			Cutoff:   c.flags.CutoffTime,
			User:     rp.UserSlackID,
			Task:     rp.Task,
			Title:    "Update Reminder",
			Fields: []slack.AttachmentField{
				{
					Title: "Task",
					Value: rp.Task,
				},
				{
					Title: "Update Deadline",
					Value: c.flags.Deadline + " " + c.flags.CutoffTime,
				},
			},
		},
	)
}

// FinalMessage delivers the final message.
func (c *Common) FinalMessage(m *shared.ReminderPayload, splay int64) error {
	c.log.Debug().
		Int64("splay", splay).
		Str("user", m.UserName).
		Msg("sleeping before sending final message")

	time.Sleep(time.Duration(splay) * time.Second)
	c.log.Debug().Msg("sleep over, sending message")

	attachment := slack.Attachment{
		Title:      m.Title,
		Text:       m.Text,
		CallbackID: c.shared.Format(shared.FinalAck),
		Color:      "#3AA3E3",
		Actions: []slack.AttachmentAction{
			{
				Name:  "accept",
				Text:  "Acknowledge",
				Type:  "button",
				Value: "ack",
			},
		},
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := c.bot.Bot().SlackClient().PostMessage(m.UserSlackID, message)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not post message")
		return err
	}

	c.log.Trace().
		Str("time", timestamp).
		Str("user", m.UserName).
		Str("channel", channelID).
		Msg("final message sent")

	return nil
}
