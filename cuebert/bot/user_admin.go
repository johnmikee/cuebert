package bot

import (
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// updateUserInfo will update information on a user when requested by an admin.
func (b *Bot) userUpdateSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values

	user, err := b.tables.UserBySlackID(ctx.Callback().View.PrivateMetadata)
	if err != nil {
		b.log.Debug().AnErr("getting user", err).Send()
		return
	}
	if user.Empty() {
		b.log.Debug().Msg("user not found")
		return
	}

	userInfo := user[0]

	original := b.Bot("Original", user)

	b.method.UserUpdateModalSubmit(&userInfo, values)

	_, err = b.tables.UpdateBRBy().
		WithOptions(userInfo).
		Parse("slack_id", ctx.Callback().View.PrivateMetadata).
		Send()
	if err != nil {
		b.log.Err(err).Msg("error updating user")
	}

	update := bot.BR{
		userInfo,
	}
	updated := b.Bot("Updated", update)

	opts := []slack.MsgOption{
		slack.MsgOptionText("User Updated", false),
		slack.MsgOptionAttachments(*original, *updated),
	}

	_, _, err = b.bot.SlackClient().PostMessage(ctx.Callback().View.ExternalID, opts...)
	if err != nil {
		b.log.Err(err).Msg("error responding")
	}

}

// userSelector will display a list of users to select from.
func (b *Bot) userSelector(triggerID, channel string) {
	headerText := slack.NewTextBlockObject(slack.MarkdownType, "Select User", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	selectUser := slack.NewTextBlockObject(slack.PlainTextType, "User", false, false)
	userOptBlock := slack.NewOptionsSelectBlockElement(slack.OptTypeUser, nil, UsersSelect)
	testingUsersBlock := slack.NewInputBlock(UpdateUser, selectUser, nil, userOptBlock)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			testingUsersBlock,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:            slack.ViewType(slack.VTModal),
		Title:           titleText,
		Close:           closeText,
		Submit:          submitText,
		Blocks:          blocks,
		CallbackID:      UpdateUserSelector,
		ClearOnClose:    false,
		NotifyOnClose:   false,
		PrivateMetadata: triggerID,
		ExternalID:      channel,
	}

	vr, err := b.bot.SlackClient().OpenView(triggerID, modalRequest)
	if err != nil {
		b.log.Error().Err(err).Msg("error opening modal")
	}

	b.log.Trace().Interface("view_response", vr).Msg("modal opened")
}

func (b *Bot) wantUpdateUser(ch string) {
	b.modalGateway(
		&modalGateway{
			text:       "Do you want to update a users values?",
			callbackID: UpdateUserQuestion,
			yesName:    UpdateUserYes,
			yesText:    "Yes",
			yesValue:   UpdateUserYes,
			yesStyle:   "primary",
			noName:     UpdateUserNo,
			noText:     "No",
			noValue:    UpdateUserNo,
			noStyle:    "danger",
			channel:    ch,
			msg:        "request to update user",
		},
	)
}
