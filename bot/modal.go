package bot

import (
	"github.com/slack-go/slack"
)

// ModalGateway is a struct to hold the information needed to send a request to invoke a modal with two buttons.
type ModalGateway struct {
	Text       string
	Fallback   string
	CallbackID string
	YesName    string
	YesText    string
	YesValue   string
	YesStyle   string
	NoName     string
	NoText     string
	NoValue    string
	NoStyle    string
	Channel    string
	Msg        string
}

// ModalGateway is a helper function to send a request to invoke a modal with two buttons.
func (b *Bot) ModalGateway(m *ModalGateway) {
	attachment := slack.Attachment{
		Text:       m.Text,
		Fallback:   m.Fallback,
		CallbackID: m.CallbackID,
		Color:      "#3AA3E3",
		Actions: []slack.AttachmentAction{
			{
				Name:  m.YesName,
				Text:  m.YesText,
				Type:  "button",
				Value: m.YesValue,
				Style: m.YesStyle,
			},
			{
				Name:  m.NoName,
				Text:  m.NoText,
				Type:  "button",
				Value: m.NoValue,
				Style: m.NoStyle,
			},
		},
	}

	b.sendAttachment(attachment, m.Channel, m.Msg)
}

func (b *Bot) sendAttachment(attachment slack.Attachment, channel, msg string) {
	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := b.bot.SlackClient().PostMessage(channel, message)
	if err != nil {
		b.log.Err(err).Msg("posting message")
	}

	b.log.Trace().
		Str("message", msg).
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Msg("message sent")
}

// CreateOptionBlockObjects - utility function for generating option block objects
func CreateOptionBlockObjects(options []string) []*slack.OptionBlockObject {
	optionBlockObjects := make([]*slack.OptionBlockObject, 0, len(options))

	for _, o := range options {
		optionText := slack.NewTextBlockObject(slack.PlainTextType, o, false, false)
		optionBlockObjects = append(optionBlockObjects, slack.NewOptionBlockObject(o, optionText, nil))
	}

	return optionBlockObjects
}
