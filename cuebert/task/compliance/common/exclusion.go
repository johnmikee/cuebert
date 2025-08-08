package common

import (
	"fmt"
	"strconv"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/db/shared/exclusions"
	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// IsExcluded implements task.Rabbit.
func (c *Common) IsExcluded(email string) (bool, bool) {
	ei, err := c.tables.Exclusion().By().Opts(exclusions.UserEmail, email).Query()
	if err != nil {
		return false, false
	}
	if ei.Empty() {
		return false, false
	}

	if ei[0].UserEmail == email {
		return true, ei[0].Approved
	}
	return false, false
}

// ExclusionYes is the helper function to request an exclusion.
func (c *Common) ExclusionYes(ctx *slacker.InteractionContext) {
	c.shared.ExclusionRequest(
		&shared.ExclusionRequest{
			TriggerID:  ctx.Callback().TriggerID,
			Reason:     "Why do you need an exclusion?",
			ReasonHint: "ex: I simply cannot update at this time.",
			Blocks:     nil,
		},
	)
}

// ExclusionRequestDecision handles the decision to approve or deny an exclusion.
func (c *Common) ExclusionRequestDecision(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0].Value

	vals := ctx.Callback().OriginalMessage.Attachments[0].Fields

	var (
		reason string
		until  string
	)

	for _, val := range vals {
		switch val.Title {
		case "Reason":
			reason = val.Value
		case "Until":
			until = val.Value
		}
	}

	ts, err := time.Parse("2006-01-02", until)
	if err != nil {
		c.log.Debug().
			Str("date", until).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	switch action {
	case c.shared.Format(shared.ApproveExclusion):
		c.log.Info().Msgf("%s - approving exclusion", ctx.Callback().User.ID)

		_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

		ui, err := c.tables.User().By().Opts(users.UserSlackID, ctx.Callback().User.ID).Query()
		if err != nil {
			c.log.Err(err).Msg("could not get user")
			return
		}
		if ui.Empty() {
			c.log.Info().Msg("no user found")
			return
		}
		_, err = c.tables.Exclusion().Add().WithOptions(
			exclusions.Info{
				Approved:  true,
				Reason:    reason,
				UpdatedAt: helpers.UpdateTime(),
				UserEmail: ui[0].UserEmail,
				Until:     ts,
			},
		).Execute()
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not approve exclusion")
		}

		_, _, err = c.bot.Bot().SlackClient().
			PostMessage(ctx.Callback().Channel.ID,
				slack.MsgOptionText("Your request for an exclusion has been approved :white_check_mark:", false))
		if err != nil {
			c.log.Err(err).Msg("posting exclusion approval")
		}

	case c.shared.Format(shared.DenyExclusion):
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

// ExclusionAddBlocks allows for custom blocks to be added to the exclusion modal.
func (c *Common) ExclusionAddBlocks() []*slack.InputBlock {
	serialSection := slack.NewTextBlockObject(slack.PlainTextType, "Serial Number", false, false)
	serialHint := slack.NewTextBlockObject(slack.PlainTextType, "ex: ABC123", false, false)
	serialInput := slack.NewPlainTextInputBlockElement(serialHint, shared.SerialInput)
	serialBlock := slack.NewInputBlock(shared.SerialNumber, serialSection, nil, serialInput)

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

// ExclusionSubmit handles the submission of the exclusion request by the admin and extracts the values from the modal.
func (c *Common) ExclusionSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	reason := values[shared.ExclusionReason][shared.ExclusionInput].Value

	c.log.Debug().
		Str("user_id", ctx.Callback().User.ID).
		Str("reason", reason).
		Msg("exlusion request")

	// get the user email from the slack id
	user, err := c.tables.User().By().Opts(users.UserSlackID, ctx.Callback().User.ID).Query()
	if err != nil {
		c.log.Err(err).Msg("getting user")
	}
	if user.Empty() {
		c.log.Info().Msg("no user found")
		return
	}

	_, err = c.tables.Exclusion().Add().WithOptions(
		exclusions.Info{
			Approved:  false,
			Reason:    reason,
			UserEmail: user[0].UserEmail,
			UpdatedAt: helpers.UpdateTime(),
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
func (c *Common) ExclusionApprove(user, reason, until string) {
	attachment := slack.Attachment{
		Title:      fmt.Sprintf("<@%s> is requesting an exclusion.", user),
		CallbackID: c.shared.Format(shared.ExclusionApprover),
		Color:      "#3AA3E3",
		Fields: []slack.AttachmentField{
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
				Name:  shared.ApproveExclusion,
				Text:  "Approve",
				Type:  "button",
				Value: shared.ApproveExclusion,
				Style: "primary",
			},
			{
				Name:  shared.DenyExclusion,
				Text:  "Deny",
				Type:  "button",
				Value: shared.DenyExclusion,
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

// Exclusions returns a list of exclusions for the given email.
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

// UserExclusionSubmit handles after the user submission once the user selected the devices they want to exclude, grabbing the values to send to the db.
func (c *Common) UserExclusionSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	dv := values[shared.ExclusionDatePicker][shared.DatePicker].SelectedDate
	reason := values[shared.ExclusionReason][shared.ExclusionInput].Value

	user := ctx.Callback().User.ID
	c.log.Trace().
		Str("user", user).
		Str("date", dv).
		Str("reason", reason).
		Msg("exclusion request")

	ts, err := time.Parse("2006-01-02", dv)
	if err != nil {
		c.log.Debug().
			Str("date", dv).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	err = c.tables.Request(ctx.Callback().User.ID, reason, ts)
	if err != nil {
		c.log.Err(err).Msg("adding exclusion request to db")
	}

	_, _, err = c.bot.Bot().SlackClient().PostMessage(ctx.Callback().User.ID,
		slack.MsgOptionText("Your request has been submitted :white_check_mark:", false))
	if err != nil {
		c.log.Err(err).Msg("posting ack for exclusion request")
	}

	c.ExclusionApprove(user, reason, dv)
}
