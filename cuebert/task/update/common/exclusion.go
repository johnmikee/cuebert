package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/update/db/devices"
	"github.com/johnmikee/cuebert/db/shared/exclusions"
	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

func (c *Common) ExclusionYes(ctx *slacker.InteractionContext) {
	di, err := c.tables.Exclusion().By().Opts(exclusions.SerialNumber, ctx.Callback().User.ID).Query()
	if err != nil {
		c.log.Err(err).Msg("could not get exclusions")
		return
	}
	c.log.Debug().Str("user", ctx.Callback().User.ID).Msg("wants to request an exclusion")

	if di.Empty() {
		_, _, err := c.bot.Bot().SlackClient().PostMessage(ctx.Callback().User.ID,
			slack.MsgOptionText("You have no devices to request an exclusion for", false))
		if err != nil {
			c.log.Err(err).Msg("posting exclusion denial")
		}
		return
	}

	// TODO: fix this
	// var devices []string
	// for _, d := range di {
	// 	devices = append(devices, d.SerialNumber)
	// }

	// c.shared.ExclusionRequest(
	// 	&shared.ExclusionRequest{
	// 		TriggerID:  ctx.Callback().TriggerID,
	// 		Reason:     "Why do you need an exclusion?",
	// 		ReasonHint: "ex: I left my computer on the moon",
	// 		Blocks:     c.exclusionRequestBlocks(devices),
	// 	},
	// )
}

// exclusionRequestDecision handles the decision to approve or deny an exclusion
func (c *Common) ExclusionRequestDecision(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0].Value

	vals := ctx.Callback().OriginalMessage.Attachments[0].Fields

	var (
		reason  string
		until   string
		serials string
	)

	for _, val := range vals {
		switch val.Title {
		case "Serial Numbers":
			serials = val.Value
		case "Reason":
			reason = val.Value
		case "Until":
			until = val.Value
		}
	}

	serialSlice := strings.Split(serials, ",")

	ts, err := time.Parse("2006-01-02", until)
	if err != nil {
		c.log.Debug().
			Str("date", until).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	switch action {
	case c.shared.Format(ApproveExclusion):
		c.log.Info().Msgf("%s - approving exclusion", ctx.Callback().User.ID)

		_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

		for _, serial := range serialSlice {
			info, err := c.tables.Device().By().Opts(devices.SerialNumber, strings.Trim(serial, " ")).Query()
			if err != nil {
				c.log.Info().Str("serial", serial).Msg("could not get info")
				continue
			}

			if info.Empty() {
				c.log.Info().Str("serial", serial).Msg("no device found ")
				continue
			}
			email := info[0].User

			slackid, err := c.tables.User().By().Opts(users.UserEmail, email).Query()
			if err != nil {
				c.log.Info().AnErr("error", err).Str("email", email).Msg("could not get slackid by email")
				continue
			}

			_, err = c.tables.Exclusion().Add().WithOptions(
				exclusions.Info{
					Approved: true,
					// SerialNumber: serial,
					Reason:    reason,
					UpdatedAt: helpers.UpdateTime(),
					UserEmail: email,
					Until:     ts,
				},
			).Execute()
			if err != nil {
				c.log.Info().AnErr("error", err).Str("serial", serial).Msg("could not approve exclusion")
			}

			_, _, err = c.bot.Bot().SlackClient().
				PostMessage(slackid[0].UserSlackID,
					slack.MsgOptionText("Your request for an exclusion has been approved :white_check_mark:", false))
			if err != nil {
				c.log.Err(err).Msg("posting exclusion approval")
			}
		}

	case c.shared.Format(DenyExclusion):
		_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

		_, _, err = c.bot.Bot().SlackClient().
			PostMessage(
				ctx.Callback().User.ID,
				slack.MsgOptionText(
					"Your request for an exclusion has been denied :octagonal_sign:",
					false,
				),
			)
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("posting exclusion denial")
		}

	default:
		c.log.Debug().Msg("got an unknown response from exclusion approval")

		_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
	}
}

func (c *Common) ExclusionAddBlocks() []*slack.InputBlock {
	serialSection := slack.NewTextBlockObject(slack.PlainTextType, "Serial Number", false, false)
	serialHint := slack.NewTextBlockObject(slack.PlainTextType, "ex: ABC123", false, false)
	serialInput := slack.NewPlainTextInputBlockElement(serialHint, SerialInput)
	serialBlock := slack.NewInputBlock(SerialNumber, serialSection, nil, serialInput)

	return []*slack.InputBlock{
		serialBlock,
	}
}

// func (c *Common) exclusionRequestBlocks(devices []string) []*slack.InputBlock {
// 	checkBoxSection := slack.NewTextBlockObject(slack.PlainTextType, "Which Device?", false, false)
// 	checkBox := slack.NewCheckboxGroupsBlockElement(DeviceBox, bot.CreateOptionBlockObjects(devices)...)
// 	checkboxBlock := slack.NewInputBlock(UserDevices, checkBoxSection, nil, checkBox)

// 	return []*slack.InputBlock{
// 		checkboxBlock,
// 	}
// }

// exclusionSubmit handles the submission of the exclusion request by the admin
// and extracts the values from the modal.
func (c *Common) ExclusionSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	dv := values[ExclusionDatePicker][DatePicker].SelectedDate
	serial := values["serial_number"]["serial_input"].Value
	reason := values[ExclusionReason][ExclusionInput].Value

	c.log.Debug().
		Str("user_id", ctx.Callback().User.ID).
		Str("serial", serial).
		Str("until", dv).
		Str("reason", reason).
		Msg("exlusion request")

	ts, err := time.Parse("2006-01-02", dv)
	if err != nil {
		c.log.Debug().
			Str("date", dv).
			Str("serial", serial).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	_, err = c.tables.Exclusion().Add().WithOptions(
		exclusions.Info{
			Approved: true,
			// SerialNumber: serial,
			Reason:    reason,
			UpdatedAt: helpers.UpdateTime(),
			Until:     ts,
		},
	).Execute()
	if err != nil {
		c.log.Err(err).Str("adding exclusion", "failed").Send()
	}

	_, _, err = c.bot.Bot().SlackClient().
		PostMessage(
			ctx.Callback().User.ID,
			slack.MsgOptionText(
				"exclusion has been submitted :white_check_mark:",
				false,
			),
		)
	if err != nil {
		c.log.Err(err).Msg("posting ack for exclusion request")
	}
}

// ExclusionApprove allows those in the approver channel to approve or deny the exclusion.
func (c *Common) ExclusionApprove(user, reason, until string, serials []string) {
	attachment := slack.Attachment{
		Title:      fmt.Sprintf("<@%s> is requesting an exclusion.", user),
		CallbackID: c.shared.Format(ExclusionApprover),
		Color:      "#3AA3E3",
		Fields: []slack.AttachmentField{
			{
				Title: "Serial Numbers",
				Value: strings.Join(serials, ", "),
			},
			{
				Title: "Reason",
				Value: reason,
			},
			{
				Title: "Until",
				Value: until,
			},
		},
		Actions: []slack.AttachmentAction{
			{
				Name:  ApproveExclusion,
				Text:  "Approve",
				Type:  "button",
				Value: ApproveExclusion,
				Style: "primary",
			},
			{
				Name:  DenyExclusion,
				Text:  "Deny",
				Type:  "button",
				Value: DenyExclusion,
				Style: "danger",
			},
		},
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := c.bot.Bot().SlackClient().PostMessage(c.config.SlackAlertChannel, message)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("posting message")
	}

	c.log.Debug().
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Msg("exclusion approval message sent")
}

func (c *Common) Exclusions(email string) ([]slack.Attachment, error) {
	user, err := c.tables.Exclusion().By().Opts(exclusions.UserEmail, email).Query()
	if err != nil {
		return nil, err
	}

	if user.Empty() {
		return nil, fmt.Errorf("no exclusions found for %s", email)
	}

	var attch slack.Attachment
	var attachments []slack.Attachment
	if !user.Empty() {
		for _, u := range user {
			attch = slack.Attachment{
				Title: "exception Info",
				Fields: []slack.AttachmentField{
					{
						Title: "Approved",
						Value: strconv.FormatBool(u.Approved),
					},
					// {
					// 	Title: "Serial Number",
					// 	Value: u.SerialNumber,
					// },
					{
						Title: "Reason",
						Value: u.Reason,
					},
					{
						Title: "Updated At",
						Value: u.UpdatedAt.Format(time.RFC3339),
					},
					{
						Title: "User Email",
						Value: u.UserEmail,
					},
					{
						Title: "Until",
						Value: u.Until.Format(time.RFC3339),
					},
				},
			}
			attachments = append(attachments, attch)
		}
		return attachments, nil
	}
	attch.Title = "No exclusions Found"
	attachments = append(attachments, attch)
	return attachments, nil
}

// after the user has selected the devices they want to exclude, this function grabs the values to send to the db.
func (c *Common) UserExclusionSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	dv := values[ExclusionDatePicker][DatePicker].SelectedDate
	reason := values[ExclusionReason][ExclusionInput].Value

	serials := exclusionSubmitSerials(ctx.Callback())
	user := ctx.Callback().User.ID
	c.log.Trace().
		Str("user", user).
		Str("date", dv).
		Str("reason", reason).
		Strs("serials", serials).
		Msg("exclusion request")

	ts, err := time.Parse("2006-01-02", dv)
	if err != nil {
		c.log.Debug().
			Str("date", dv).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	err = c.tables.Request(ctx.Callback().User.ID, reason, serials, ts)
	if err != nil {
		c.log.Err(err).Msg("adding exclusion request to db")
	}

	_, _, err = c.bot.Bot().SlackClient().PostMessage(ctx.Callback().User.ID,
		slack.MsgOptionText("Your request has been submitted :white_check_mark:", false))
	if err != nil {
		c.log.Err(err).Msg("posting ack for exclusion request")
	}

	c.ExclusionApprove(user, reason, dv, serials)
}

func exclusionSubmitSerials(callback *slack.InteractionCallback) []string {
	serials := []string{}
	for _, v := range callback.View.State.Values["user_devices"]["device_box"].SelectedOptions {
		serials = append(serials, v.Text.Text)
	}
	return serials
}
