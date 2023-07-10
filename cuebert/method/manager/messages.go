package manager

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/slack-go/slack"
)

func firstMessage(day time.Time) string {
	return fmt.Sprintf(`
Hello, you are receiving this message because your laptop macOS is out of date.

In order to have continued access to Ramp Data/Systems (e.g., Gmail, Okta, Zoom) your device must be compliant with our company <https://app.drata.com/employee/policy/30|security policies>. 
Our policies *state that your macOS must be up to date* because upgrading your device is crucial for a secure work environment.

If your device continues to stay out of compliance, you will lose access to Ramp data/systems at the end of the week.

To upgrade macOS, go to *System Preferences*, and click *Software Update*.

Once you have clicked *Upgrade Now*, the update will begin downloading.  A progress bar will show the status of the download and during this time you can still use your computer as you normally would.

*Your Manager will be engaged if your device is not compliant by %s.*

Post in #team-it if you have any trouble updating your machine.
	`, day.Format("Monday, January 2, 2006"))
}

func managerMessaging(userName, firstMessageSent, usersSlackID string) string {
	return fmt.Sprintf(`
We would like to bring to your attention that %s has not yet upgraded their laptop to the latest operating system. We sent previous communication to do so on %s.

As %s manager, please work with them to ensure they are not locked out of Ramp Data/Systems (e.g., Gmail, Slack, Zoom) by having them update their macOS by EOW.

Ramp issued devices not in compliance according to our <https://app.drata.com/employee/policy/30|Information Security Policies> will have access to Ramp data/systems restricted by the end of the week.

<@%s>, please collaborate with your manager to complete the upgrade promptly.

For more information and detailed guidance, refer to this <https://ramp.slab.com/posts/os-upgrade-guidance-bx3mz5qp|Slab Article>.
`,
		userName,
		firstMessageSent,
		helpers.PossessiveForm(userName),
		usersSlackID,
	)
}

// currently not implemented
func (m *Manager) ReminderMessage(rp *bot.ReminderPayload) error {
	return nil
}

func (m *Manager) managerMessage(rp *bot.ReminderPayload) {
	attachment := slack.Attachment{
		Text: managerMessaging(
			rp.UserName,
			rp.FirstMessage,
			rp.UserSlackID,
		),
		CallbackID: GroupDM,
		Color:      "#3AA3E3",
	}

	message := slack.MsgOptionAttachments(attachment)

	dm, _, _, err := m.sc.OpenConversation(
		&slack.OpenConversationParameters{
			ChannelID: "",
			ReturnIM:  false,
			Users: []string{
				rp.ManagerSlackID,
				rp.UserSlackID,
			},
		},
	)
	if err != nil {
		m.log.Err(err).Msg("creating dm with manager")
		return
	}

	channelID, timestamp, err := m.sc.PostMessage(dm.ID, message)
	if err != nil {
		m.log.Err(err).Msg("posting message in")
	}

	m.log.Debug().
		Str("serial", rp.Serial).
		Str("time", timestamp).
		Str("user", rp.UserName).
		Str("channel", channelID).
		Msg("manager message sent")

	err = m.tables.ManagerNotifed(true, rp.Serial)
	if err != nil {
		m.log.Err(err).Send()
	}
}
