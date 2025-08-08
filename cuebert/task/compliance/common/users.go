package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/johnmikee/cuebert/cuebert/task/info"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

var (
	titleText  = slack.NewTextBlockObject(slack.PlainTextType, "Cuebert", false, false)
	closeText  = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	submitText = slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
	yes        = slack.NewTextBlockObject(slack.PlainTextType, "Yes", false, false)
	no         = slack.NewTextBlockObject(slack.PlainTextType, "No", false, false)
)

// UserCompliant implements shared.Caller checking if a user is compliant.
func (c *Common) UserCompliant(slackID string) error {
	ui, err := c.tables.User().By().Opts(users.UserSlackID, slackID).Query()
	if err != nil {
		return err
	}

	if ui.Empty() {
		return fmt.Errorf("no user found for slack id %s", slackID)
	}

	c.log.Info().Interface("ui", ui).Msg("ui")

	return c.idp.RemoveFromBlocked(ui[0].UserEmail)
}

// AddAllUsers will take the users from the compliance task implementation and add them to the DB.
func (c *Common) AddAllUsers(cu []info.Users) ([]string, error) {
	ui := []users.Info{}

	slackUsers, err := c.bot.Bot().SlackClient().GetUsers()
	if err != nil {
		return nil, err
	}
	// now we make a map of the slack users for quick lookup.
	// the key is the email and the id and user name are the values
	slackMap := make(map[string]slack.User)
	for i := range slackUsers {
		slackMap[slackUsers[i].Profile.Email] = slackUsers[i]
	}

	missed := []string{}
	for i := range cu {
		c.log.Info().Interface("cu", cu[i]).Msg("cu")
		// now match the compliance users by email to the slack users
		// and combine the info to build the users.Info struct
		if slackUser, ok := slackMap[cu[i].Email]; ok {
			ui = append(ui, users.Info{
				UserLongName: slackUser.RealName,
				UserEmail:    cu[i].Email,
				UserSlackID:  slackUser.ID,
				TZOffset:     int64(slackUser.TZOffset),
			})
		} else {
			missed = append(missed, cu[i].Email)
		}
		for _, m := range missed {
			su, err := c.bot.Bot().SlackClient().GetUserByEmail(m)
			if err != nil {
				c.log.Debug().AnErr("getting user", err).Send()
				continue
			}
			ui = append(ui,
				users.Info{
					UserLongName: su.RealName,
					UserEmail:    m,
					UserSlackID:  su.ID,
					TZOffset:     int64(su.TZOffset),
				},
			)
		}
	}

	_, err = c.tables.User().Add().All(ui)
	if err != nil {
		return nil, err
	}

	return nil, err
}

// UserUpdate is the handler for the user update command.
func (c *Common) UserUpdate(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	userID := values[shared.UpdateUser][shared.UsersSelect].SelectedUser

	user, err := c.tables.Tracking().By().Opts(tracking.UserSlackID, userID).Query()
	if err != nil {
		c.log.Debug().AnErr("getting user", err).Send()
		return
	}
	if user.Empty() {
		c.log.Debug().Msg("user not found")
		return
	}

	userInfo := user[0]

	headerText := slack.NewTextBlockObject(slack.MarkdownType, "Update User", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// now we load the values from the user and set them as the default values for the input blocks
	email := slack.NewTextBlockObject(slack.PlainTextType, "Email", false, false)
	emailPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.UserEmail, false, false)
	emailInput := slack.NewPlainTextInputBlockElement(emailPlaceHolder, "email")
	emailBlock := slack.NewInputBlock(shared.Email, email, nil, emailInput)
	emailBlock.Optional = true

	firstAck := slack.NewTextBlockObject(slack.PlainTextType, "First Message Acknowledged", false, false)
	firstAckSentYes := slack.NewOptionBlockObject(shared.TruthyYes, yes, nil)
	firstAckSentNo := slack.NewOptionBlockObject(shared.TruthyNo, no, nil)
	firstAckOptions := slack.NewRadioButtonsBlockElement(shared.FirstMessageACKOption, firstAckSentYes, firstAckSentNo)
	firstAckBlock := slack.NewInputBlock(shared.FirstMessageACK, firstAck, nil, firstAckOptions)
	firstAckBlock.Optional = true

	firstAckAt := slack.NewTextBlockObject(slack.PlainTextType, "First Message Acknowledged At", false, false)
	firstAckAtPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.FirstACKTime.Format(time.RFC3339), false, false)
	firstAckAtDateBlock := slack.NewDatePickerBlockElement(shared.FirstMessageACKAt)
	firstAckAtInput := slack.NewInputBlock(shared.FirstMessageACKAtPicker, firstAckAt, firstAckAtPlaceHolder, firstAckAtDateBlock)
	firstAckAtInput.Optional = true

	firstSent := slack.NewTextBlockObject(slack.PlainTextType, "First Message Sent", false, false)
	firstSentYes := slack.NewOptionBlockObject(shared.TruthyYes, yes, nil)
	firstSentNo := slack.NewOptionBlockObject(shared.TruthyNo, no, nil)
	firstSentOptions := slack.NewRadioButtonsBlockElement(shared.FirstMessageSentOption, firstSentYes, firstSentNo)
	firstSentBlock := slack.NewInputBlock(shared.FirstMessageSentOption, firstSent, nil, firstSentOptions)
	firstSentBlock.Optional = true

	firstSentAt := slack.NewTextBlockObject(slack.PlainTextType, "First Message Sent At", false, false)
	firstSentAtPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.FirstMessageSentAt.Format(time.RFC3339), false, false)
	firstSentAtDateBlock := slack.NewDatePickerBlockElement(shared.FirstMessageSentAt)
	firstSentAtInput := slack.NewInputBlock(shared.FirstMessageSentAtPicker, firstSentAt, firstSentAtPlaceHolder, firstSentAtDateBlock)
	firstSentAtInput.Optional = true

	tzOffset := slack.NewTextBlockObject(slack.PlainTextType, "Timezone Offset", false, false)
	tzOffsetPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, helpers.TZStringer(userInfo.TZOffset), false, false)
	tzOffsetInput := slack.NewPlainTextInputBlockElement(tzOffsetPlaceHolder, "tz_offset")
	tzOffsetBlock := slack.NewInputBlock("tz_offset", tzOffset, nil, tzOffsetInput)
	tzOffsetBlock.Optional = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			emailBlock,
			firstAckBlock,
			firstAckAtInput,
			firstSentBlock,
			firstSentAtInput,
			tzOffsetBlock,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           titleText,
		Close:           closeText,
		Submit:          submitText,
		Blocks:          blocks,
		CallbackID:      c.shared.Format(shared.UserUpdateModal),
		PrivateMetadata: userInfo.UserSlackID,
		ExternalID:      ctx.Callback().View.ExternalID,
	}

	_, err = c.bot.Bot().SlackClient().OpenView(ctx.Callback().TriggerID, modalRequest)
	if err != nil {
		c.log.Error().Err(err).Msg("error opening modal")
	}
}

// UserUpdateSubmit is the handler for the user update modal submission.
func (c *Common) UserUpdateSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values

	user, err := c.tables.Tracking().By().Opts(tracking.UserSlackID, ctx.Callback().View.PrivateMetadata).Query()
	if err != nil {
		c.log.Debug().AnErr("getting user", err).Send()
		return
	}
	if user.Empty() {
		c.log.Debug().Msg("user not found")
		return
	}

	userInfo := user[0]
	original := c.Tracking("Original", user)

	c.UserUpdateModalSubmit(&userInfo, values)

	_, err = c.tables.Tracking().Update().
		WithOptions(userInfo).
		Parse(tracking.UserSlackID, ctx.Callback().View.PrivateMetadata).
		Send()
	if err != nil {
		c.log.Err(err).Msg("error updating user")
	}

	update := tracking.TI{
		userInfo,
	}
	updated := c.Tracking("Updated", update)

	opts := []slack.MsgOption{
		slack.MsgOptionText("User Updated", false),
		slack.MsgOptionAttachments(*original, *updated),
	}

	_, _, err = c.bot.Bot().SlackClient().PostMessage(ctx.Callback().View.ExternalID, opts...)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("responding")
	}
}

// UserUpdateModalSubmit extracts the values from the modal and formats as needed.
func (c *Common) UserUpdateModalSubmit(base *tracking.Info, values map[string]map[string]slack.BlockAction) {
	if values[shared.Email][shared.Email].Value != "" {
		base.UserEmail = values[shared.Email][shared.Email].Value
	}
	if values[shared.FirstMessageACK][shared.FirstMessageACKOption].SelectedOption.Value != "" {
		base.FirstACK = helpers.StringBooler(values[shared.FirstMessageACK][shared.FirstMessageACKOption].SelectedOption.Value)
	}
	if values[shared.FirstMessageACKAtPicker][shared.FirstMessageACKAt].SelectedOption.Value != "" {
		ts, _ := helpers.StringToTime(values[shared.FirstMessageACKAtPicker][shared.FirstMessageACKAt].SelectedDate)
		base.FirstACKTime = ts
	}
	if values[shared.FirstMessageSentOption][shared.FirstMessageSentOption].SelectedOption.Value != "" {
		base.FirstMessageSent = helpers.StringBooler(values[shared.FirstMessageSentAt][shared.FirstMessageSentAt].SelectedOption.Value)
	}
}

// GetUsersInfo is the admin command the query a user and return their information in slack.
func (c *Common) GetUsersInfo() {
	reportOpts := []string{"slackid", "email"}
	definition := &slacker.CommandDefinition{
		Command:     "get users info <opt> {input}",
		Description: "Get info about a user",
		Examples:    []string{"get users info slackid <id>", "get users info email <email>"},
		Middlewares: []slacker.CommandMiddlewareHandler{ops.AuthorizationMiddleware(c.flags.AuthUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			opt := ctx.Request().Param("opt")
			which := ctx.Request().Param("input")

			attachments := []slack.Attachment{
				{Color: "blue", AuthorName: "cuebert"},
			}

			var (
				user  tracking.TI
				err   error
				email string
			)

			switch strings.ToLower(opt) {
			case "slackid":
				user, err = c.tables.Tracking().By().Opts(tracking.UserSlackID, which).Query()
			case "email":
				which = helpers.ExtractEmails(which)
				c.log.Trace().Str("which", which).Msg("email")
				user, err = c.tables.Tracking().By().Opts(tracking.UserEmail, which).Query()
				c.log.Info().Interface("user", user).Msg("user")
			default:
				msg := ops.FuzzyMatchNonOpt(opt, reportOpts)
				_, err := ctx.Response().Reply(msg)
				if err != nil {
					c.log.Debug().AnErr("sending response", err).Send()
				}
				return
			}

			if err != nil {
				c.log.Debug().AnErr("getting user", err).Send()
			}

			if !user.Empty() {
				email = user[0].UserEmail
				ui, err := c.User("User Info", email)
				if err != nil {
					c.log.Info().AnErr("error", err).Msg("error getting user")
				}
				attachments = append(attachments, *ui)

				bi := c.Tracking("Tracking Results Table", user)
				attachments = append(attachments, *bi)

				ei, err := c.Exclusions(email)
				if err != nil {
					c.log.Err(err).Msg("error getting exclusions")
				}
				attachments = append(attachments, ei...)

			}

			_, err = ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
			if err != nil {
				c.log.Err(err).Msg("error responding")
			}
		},
	}

	c.bot.Bot().AddCommand(definition)
}
