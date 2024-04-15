package shared

import (
	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

var (
	TitleText  = slack.NewTextBlockObject(slack.PlainTextType, "Cuebert", false, false)
	CloseText  = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	SubmitText = slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
)

// ComplianceChecker runs off user interaction to check compliance status.
func (c *Shared) ComplianceChecker(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0]
	_, _, err := c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not delete the compliance request message")
	}

	if action.Name == YesCheck {
		ok, err := c.impl.CheckComplianceStatus(ctx.Callback().User.ID)
		c.log.Info().Bool("compliant", ok).AnErr("error", err).Msg("checking compliance status")
		var message string
		if ok {
			c.log.Info().Msg("user is now compliant")
			message = "You are now compliant! :white_check_mark:"
			err = c.caller.UserCompliant(ctx.Callback().User.ID)
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("could not update user compliance")
				message += " :alert: There was an error updating your compliance status."
			}
		} else {
			message = ":alert: You have not completed the required training. Please complete the training and try again."
			if c.flags.TaskImplementation == string(implementation.OS) {
				message = ":alert: You have not completed the required update. Please update your device and try again."
			}
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("could not check compliance status")
				switch err.Error() {
				case "no user found", "no_text":
					message = ":white_check_mark:"
				default:
					c.log.Info().Msg("user is not compliant")
				}
			}
		}

		c.log.Trace().Str("text", message).Msg("sending message")

		_, _, err = c.bot.Bot().SlackClient().PostMessage(ctx.Callback().User.ID,
			slack.MsgOptionText(message, false))
		if err != nil {
			c.log.Err(err).Msg("posting message")
		}
	}
}

// userSelector will display a list of users to select from.
func (c *Shared) UserSelector(triggerID, channel string) {
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
		Title:           TitleText,
		Close:           CloseText,
		Submit:          SubmitText,
		Blocks:          blocks,
		CallbackID:      c.constFormat(UpdateUserSelector),
		ClearOnClose:    false,
		NotifyOnClose:   false,
		PrivateMetadata: triggerID,
		ExternalID:      channel,
	}

	vr, err := c.bot.Bot().SlackClient().OpenView(triggerID, modalRequest)
	if err != nil {
		c.log.Error().Err(err).Msg("error opening modal")
	}

	c.log.Trace().Interface("view_response", vr).Msg("modal opened")
}

func (c *Shared) WantUpdateUser(ch string) {
	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       "Do you want to update a users values?",
			CallbackID: UpdateUserQuestion,
			YesName:    UpdateUserYes,
			YesText:    "Yes",
			YesValue:   UpdateUserYes,
			YesStyle:   "primary",
			NoName:     UpdateUserNo,
			NoText:     "No",
			NoValue:    UpdateUserNo,
			NoStyle:    "danger",
			Channel:    ch,
			Msg:        "request to update user",
		},
	)
}

// updateUserInfoInteractive allows authorized users to set the initial configuration for the bot.
func (c *Shared) UpdateUserInfoInteractive() {
	definition := &slacker.CommandDefinition{
		Command:     "update user interactive",
		Description: "Update info about a user",
		Examples:    []string{"update user interactive"},
		Middlewares: []slacker.CommandMiddlewareHandler{ops.AuthorizationMiddleware(c.flags.AuthUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			c.WantUpdateUser(ctx.Event().ChannelID)
		},
	}

	c.bot.Bot().AddCommand(definition)
}
