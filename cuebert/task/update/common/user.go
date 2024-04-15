package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

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

func (c *Common) GetUserTZ(slackID string) (int64, error) {
	ui, err := c.tables.User().By().Opts(users.UserSlackID, slackID).Query()
	if err != nil {
		return 0, err
	}

	if ui.Empty() {
		return 0, fmt.Errorf("no user found for slack id %s", slackID)
	}
	return ui[0].TZOffset, nil
}

// GetSelfInfo returns information about the user when requested.
func (c *Common) GetSelfInfo() {
	definition := &slacker.CommandDefinition{
		Command:     "get my info",
		Description: "Get info about your account",
		Examples:    []string{"get my info"},
		Handler: func(ctx *slacker.CommandContext) {
			attachments := []slack.Attachment{
				{Color: "blue", AuthorName: "cuebert"},
			}
			user, err := c.tables.User().By().Opts(users.UserSlackID, ctx.Event().UserID).Query()
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("error getting user")
				attachments = append(attachments, slack.Attachment{Color: "red", Text: "error getting user"})

				_, err := ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
				if err != nil {
					c.log.Err(err).Msg("error responding")
				}
				return
			}

			if !user.Empty() {
				email := user[0].UserEmail
				ui, err := c.User("User Info", email)
				if err != nil {
					c.log.Info().AnErr("error", err).Msg("error getting user")
				}
				attachments = append(attachments, *ui)
			}

			_, err = ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("error responding")
			}
		},
	}
	c.bot.Bot().AddCommand(definition)
}

// GetMDMUsers will return all users from the MDM
func (c *Common) GetMDMUsers(opts *mdm.QueryOpts) ([]mdm.User, error) {
	var res mdm.DeviceResults
	var err error

	if opts != nil {
		res, err = c.withQuery(opts)
	} else {
		res, err = c.all()
	}

	if err != nil {
		return nil, err
	}

	mu := []mdm.User{}
	for i := range res {
		mu = append(mu, res[i].User)
	}

	return mu, nil
}

func (c *Common) all() (mdm.DeviceResults, error) {
	return c.mdm.ListDevices()
}

func (c *Common) withQuery(opts *mdm.QueryOpts) (mdm.DeviceResults, error) {
	return c.mdm.QueryDevices(opts)
}

// getUsersInfo returns information about a user when requested by an admin.
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
			c.log.Trace().Str("command", "get report").Str("opt", opt).Str("which", which).Send()

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

				bi := c.Bot("Bot Results Table", user)
				attachments = append(attachments, *bi)

				ei, err := c.Exclusions(email)
				if err != nil {
					c.log.Err(err).Msg("error getting exclusions")
				}
				attachments = append(attachments, ei...)

				di, err := c.Device(email)
				if err != nil {
					c.log.Err(err).Msg("error getting devices")
				}
				attachments = append(attachments, di...)
			}

			_, err = ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
			if err != nil {
				c.log.Err(err).Msg("error responding")
			}
		},
	}

	c.bot.Bot().AddCommand(definition)
}

func (c *Common) User(title, email string) (*slack.Attachment, error) {
	user, err := c.tables.User().By().Opts(users.UserEmail, email).Query()
	if err != nil {
		return nil, err
	}

	if user.Empty() {
		return nil, nil
	}
	var attch slack.Attachment
	for _, u := range user {
		attch = slack.Attachment{
			Title: title,
			Fields: []slack.AttachmentField{
				{
					Title: "MDM ID",
					Value: strconv.Itoa(u.UserServiceID),
				},
				{
					Title: "User Name",
					Value: u.UserLongName,
				},
				{
					Title: "User Email",
					Value: u.UserEmail,
				},
				{
					Title: "User Slack ID",
					Value: u.UserSlackID,
				},
				{
					Title: "User Timezone",
					Value: helpers.TZStringer(u.TZOffset),
				},
				{
					Title: "Updated at",
					Value: u.UpdatedAt.Format(time.RFC3339),
				},
			},
		}
	}

	return &attch, nil
}

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
	original := c.Bot("Original", user)

	c.UserUpdateModalSubmit(&userInfo, values)

	_, err = c.tables.Tracking().Update().
		WithOptions(userInfo).
		Parse("slack_id", ctx.Callback().View.PrivateMetadata).
		Send()
	if err != nil {
		c.log.Err(err).Msg("error updating user")
	}

	update := tracking.TI{
		userInfo,
	}
	updated := c.Bot("Updated", update)

	opts := []slack.MsgOption{
		slack.MsgOptionText("User Updated", false),
		slack.MsgOptionAttachments(*original, *updated),
	}

	_, _, err = c.bot.Bot().SlackClient().PostMessage(ctx.Callback().View.ExternalID, opts...)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("responding")
	}
}

func (c *Common) UserUpdate(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	userID := values[UpdateUser][UsersSelect].SelectedUser

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
	emailBlock := slack.NewInputBlock(Email, email, nil, emailInput)
	emailBlock.Optional = true

	firstAck := slack.NewTextBlockObject(slack.PlainTextType, "First Message Acknowledged", false, false)
	firstAckSentYes := slack.NewOptionBlockObject(TruthyYes, yes, nil)
	firstAckSentNo := slack.NewOptionBlockObject(TruthyNo, no, nil)
	firstAckOptions := slack.NewRadioButtonsBlockElement(FirstMessageACKOption, firstAckSentYes, firstAckSentNo)
	firstAckBlock := slack.NewInputBlock(FirstMessageACK, firstAck, nil, firstAckOptions)
	firstAckBlock.Optional = true

	firstAckAt := slack.NewTextBlockObject(slack.PlainTextType, "First Message Acknowledged At", false, false)
	firstAckAtPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.FirstACKTime.Format(time.RFC3339), false, false)
	firstAckAtDateBlock := slack.NewDatePickerBlockElement(FirstMessageACKAt)
	firstAckAtInput := slack.NewInputBlock(FirstMessageACKAtPicker, firstAckAt, firstAckAtPlaceHolder, firstAckAtDateBlock)
	firstAckAtInput.Optional = true

	firstSent := slack.NewTextBlockObject(slack.PlainTextType, "First Message Sent", false, false)
	firstSentYes := slack.NewOptionBlockObject(TruthyYes, yes, nil)
	firstSentNo := slack.NewOptionBlockObject(TruthyNo, no, nil)
	firstSentOptions := slack.NewRadioButtonsBlockElement(FirstMessageSentOption, firstSentYes, firstSentNo)
	firstSentBlock := slack.NewInputBlock(FirstMessageSentOption, firstSent, nil, firstSentOptions)
	firstSentBlock.Optional = true

	firstSentAt := slack.NewTextBlockObject(slack.PlainTextType, "First Message Sent At", false, false)
	firstSentAtPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.FirstMessageSentAt.Format(time.RFC3339), false, false)
	firstSentAtDateBlock := slack.NewDatePickerBlockElement(FirstMessageSentAt)
	firstSentAtInput := slack.NewInputBlock(FirstMessageSentAtPicker, firstSentAt, firstSentAtPlaceHolder, firstSentAtDateBlock)
	firstSentAtInput.Optional = true

	serial := slack.NewTextBlockObject(slack.PlainTextType, "Serial Number", false, false)
	serialPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, "userInfo.SerialNumber", false, false)
	serialInput := slack.NewPlainTextInputBlockElement(serialPlaceHolder, SerialNumber)
	serialBlock := slack.NewInputBlock(SerialNumber, serial, nil, serialInput)
	serialBlock.Optional = true

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
			serialBlock,
			tzOffsetBlock,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           titleText,
		Close:           closeText,
		Submit:          submitText,
		Blocks:          blocks,
		CallbackID:      c.shared.Format(UserUpdateModal),
		PrivateMetadata: userInfo.UserSlackID,
		ExternalID:      ctx.Callback().View.ExternalID,
	}

	_, err = c.bot.Bot().SlackClient().OpenView(ctx.Callback().TriggerID, modalRequest)
	if err != nil {
		c.log.Error().Err(err).Msg("error opening modal")
	}
}
