package main

import (
	"fmt"

	"github.com/slack-go/slack"
)

type managerAlert struct {
	info []missingManager
	msg  string
}

// sendAlert sends an alert to the specified channel after building the attachment
func (b *Bot) sendAlert(alertChan string, attachment *slack.Attachment) {
	_, _, err := b.bot.Client().PostMessage(alertChan, slack.MsgOptionAttachments(*attachment))
	if err != nil {
		b.log.Err(err).Msg("sending alert message")
	}
}

// buildAlert builds the attachment for the alert message
func buildAlert(source string, mm []missingManager) slack.Attachment {
	fields := []slack.AttachmentField{}
	for _, m := range mm {
		fields = append(fields, slack.AttachmentField{
			Title: m.user,
			Value: m.userEmail,
		})
	}
	attachment := slack.Attachment{
		Text:   fmt.Sprintf("The following users are missing a manager in %s", source),
		Fields: fields,
	}

	return attachment
}

// alertIfNoManager sends an alert if there are any users missing a manager.
// when this was first written, there was a logic error where users were
// missing managers in both the DB and iDP. this is no longer the case, but
// the logic is still here in case it is needed in the future to range
// over potential sources of truth for managers values.
func (b *Bot) alertIfNoManager(alertChan string, ma []managerAlert) {
	for _, m := range ma {
		alert := buildAlert(m.msg, m.info)
		b.sendAlert(alertChan, &alert)
	}
}
