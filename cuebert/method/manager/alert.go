package manager

import (
	"fmt"

	"github.com/slack-go/slack"
)

type ManagerAlert struct {
	info []MissingManager
	msg  string
}

// sendAlert sends an alert to the specified channel after building the attachment
func (m *Manager) sendAlert(alertChan string, attachment *slack.Attachment) {
	_, _, err := m.sc.PostMessage(alertChan, slack.MsgOptionAttachments(*attachment))
	if err != nil {
		m.log.Err(err).Msg("sending alert message")
	}
}

// buildAlert builds the attachment for the alert message
func buildAlert(source string, mm []MissingManager) slack.Attachment {
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
func (m *Manager) alertIfNoManager(alertChan string, ma []ManagerAlert) {
	for _, a := range ma {
		alert := buildAlert(a.msg, a.info)
		m.sendAlert(alertChan, &alert)
	}
}
