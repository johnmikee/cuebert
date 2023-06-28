package main

import (
	"fmt"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/slack-go/slack"
)

type missingManager struct {
	user      string
	userEmail string
}

// check is the list of users who were not added to the db. this means they
// did not exist in the MDM. a likely cause is that the user has a non-macOS
// machine ala windows.
//
// we take a second pass on these and attach it directly to the user.
func associateUserManager(b *Bot, check []string) {
	b.cfg.log.Info().Msg("checking manager association..")

	ur, err := b.idp.GetAllUsers()
	if err != nil {
		b.cfg.log.Trace().
			AnErr("error", err).
			Msg("getting all idp users")
		return
	}

	for i := range ur {
		b.log.Trace().
			Str("user", ur[i].Profile.Email).
			Msg("checking user")
	}

	missing, err := b.tables.buildAssociation(ur, check, b.bot.Client())
	if err != nil {
		b.log.Err(err).
			Msg("associating managers to users")
		return
	}
	b.log.Trace().Msg("alerting if no manager")

	// if we have missing managers we need to alert
	if b.cfg.flags.sendManagerMissing {
		// that is, if we opted to do so.
		b.alertIfNoManager(b.cfg.SlackAlertChannel, []managerAlert{
			{
				info: missing,
				msg:  "the db",
			},
		})
	}
}

func managerMessaging(userName, firstMessageSent, usersSlackID string) string {
	return fmt.Sprintf(`
We would like to bring to your attention that %s has not yet upgraded their laptop to the latest operating system. We sent previous communication to do so on %s.

As %s manager, please work with them to ensure they are not locked out of Megacorp Systems (e.g., Gmail, Slack, Zoom) by having them update their macOS by EOW.

Megacorp issued devices not in compliance according to our <https://megacorp.com/article|Information Security Policies> will have access to Megacorp data/systems restricted by the end of the week.

<@%s>, please collaborate with your manager to complete the upgrade promptly.

For more information and detailed guidance, refer to this <https://megacorp|Article>.
`,
		userName,
		firstMessageSent,
		helpers.PossessiveForm(userName),
		usersSlackID,
	)
}

func (b *Bot) managerMessage(rp *reminderPayload) {
	attachment := slack.Attachment{
		Text: managerMessaging(
			rp.userName,
			rp.firstMessage,
			rp.userSlackID,
		),
		CallbackID: "group_dm",
		Color:      "#3AA3E3",
	}

	message := slack.MsgOptionAttachments(attachment)

	dm, _, _, err := b.bot.Client().OpenConversation(&slack.OpenConversationParameters{
		ChannelID: "",
		ReturnIM:  false,
		Users: []string{
			rp.managerSlackID,
			rp.userSlackID,
		},
	})
	if err != nil {
		b.log.Err(err).Msg("creating dm with manager")
		return
	}

	channelID, timestamp, err := b.bot.Client().PostMessage(dm.ID, message)
	if err != nil {
		b.log.Err(err).Msg("posting message in")
	}

	b.log.Debug().
		Str("serial", rp.serial).
		Str("time", timestamp).
		Str("user", rp.userName).
		Str("channel", channelID).
		Msg("manager message sent")

	err = b.db.ManagerNotifed(true, rp.serial)
	if err != nil {
		b.log.Err(err).Send()
	}
}
