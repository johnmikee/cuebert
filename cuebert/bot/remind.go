package bot

import (
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

type ReminderInfo struct {
	Deadline string
	Cutoff   string
	User     string
	Serial   string
	Version  string
	OS       string
	Text     string
}

type ReminderPayload struct {
	UserSlackID    string
	ManagerSlackID string
	UserName       string
	Serial         string
	Model          string
	OS             string
	FirstMessage   string
	TZOffset       int64
}

// reminderHelp adds the reminder command to the bot
func (b *Bot) reminderHelp() {
	definition := &slacker.CommandDefinition{
		Command:     "request reminder",
		Description: "Request a reminder to update",
		Examples:    []string{"request reminder"},
		Handler: func(ctx *slacker.CommandContext) {
			_, err := ctx.Response().ReplyBlocks(
				[]slack.Block{
					slack.NewSectionBlock(
						slack.NewTextBlockObject("mrkdwn", "Would you like to request a reminder to update?", false, false), nil, nil),
					slack.NewActionBlock(
						RemindMeQuestion,
						slack.NewButtonBlockElement(
							RemindMe,
							RemindMe,
							slack.NewTextBlockObject("plain_text", Yes, false, false)).
							WithStyle(slack.StylePrimary),
						slack.NewButtonBlockElement(
							DontRemindMe,
							DontRemindMe,
							slack.NewTextBlockObject("plain_text", No, false, false)).
							WithStyle(slack.StyleDanger)),
				},
			)
			if err != nil {
				b.log.Debug().AnErr("sending reminder prompt", err).
					Send()
			}
		},
	}

	b.bot.AddCommand(definition)
}

// deliverReminder delivers the reminder to the user
func (b *Bot) deliverReminder(ri *ReminderInfo) error {
	attachment := slack.Attachment{
		Title:      "Cuebert Update Reminder",
		Text:       ri.Text,
		CallbackID: UserReminder,
		Color:      "#3AA3E3",
		Fields: []slack.AttachmentField{
			{
				Title: "Required Version",
				Value: ri.Version,
			},
			{
				Title: "Update Deadline",
				Value: ri.Deadline + " " + ri.Cutoff,
			},
			{
				Title: "Current Version",
				Value: ri.OS,
			},
			{
				Title: "Serial Number",
				Value: ri.Serial,
			},
		},
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := b.bot.SlackClient().PostMessage(ri.User, message)
	if err != nil {
		b.log.Err(err).Msg("posting message")
		return err
	}

	b.log.Debug().
		Str("user", ri.User).
		Str("serial", ri.Serial).
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Str("table", "bot_results").
		Bool("sent", true).
		Msg("delivered reminder")

	return nil
}
