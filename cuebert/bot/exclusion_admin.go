package bot

import (
	"fmt"
	"strings"
	"time"

	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// exclusionApprove allows those in the approver channel to approve or deny the exclusion.
func (b *Bot) exclusionApprove(user, reason, until string, serials []string) {
	attachment := slack.Attachment{
		Title:      fmt.Sprintf("<@%s> is requesting an exclusion.", user),
		CallbackID: ExclusionApprover,
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

	channelID, timestamp, err := b.bot.SlackClient().PostMessage(b.cfg.slackAlertChannel, message)
	if err != nil {
		b.log.Err(err).Msg("posting message")
	}

	b.log.Debug().
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Msg("exclusion approval message sent")
}

// exclusionRequestDecision handles the decision to approve or deny an exclusion
func (b *Bot) exclusionRequestDecision(ctx *slacker.InteractionContext) {
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
		b.log.Debug().
			Str("date", until).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	switch action {
	case "approve_exclusion":
		b.log.Info().Msgf("%s - approving exclusion", ctx.Callback().User.ID)

		_, _, _ = b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

		for _, serial := range serialSlice {
			info, err := b.tables.DeviceBySerial(strings.Trim(serial, " "))
			if err != nil {
				b.log.Err(err).Msgf("could not get serial info for %s", serial)
				continue
			}

			email := info[0].User

			slackid, err := b.tables.UserByEmail(email)

			if err != nil {
				b.log.Err(err).Msgf("could not get slackid by email %s", email)
				return
			}

			err = b.tables.ApproveExclusion(reason, serial, ts)

			if err != nil {
				b.log.Err(err).Msgf("could not approve exclusion for %s", serial)
			}

			_, _, err = b.bot.SlackClient().PostMessage(slackid[0].UserSlackID,
				slack.MsgOptionText("Your request for an exclusion has been approved :white_check_mark:", false))
			if err != nil {
				b.log.Err(err).Msg("posting exclusion approval")
			}
		}

	case "deny_exclusion":
		_, _, _ = b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

		_, _, err = b.bot.SlackClient().PostMessage(ctx.Callback().User.ID,
			slack.MsgOptionText("Your request for an exclusion has been denied :octagonal_sign:", false))
		if err != nil {
			b.log.Err(err).Msg("posting exclusion denial")
		}

	default:
		b.log.Debug().Msgf("got an unknown response from exclusion approval")

		_, _, _ = b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
	}

}

// exclusionRequest is the modal presented to the admins who requested to add
// an exclusion to the exclusion list to load the exclusion into the database.
func (b *Bot) exclusionAdd(triggerID string) {
	headerText := slack.NewTextBlockObject(slack.MarkdownType, "Add Exclusion", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	reason := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: ExclusionReason,
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Reason for Exclusion",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.PlainTextInputBlockElement{
			Type:     slack.METPlainTextInput,
			ActionID: ExclusionInput,
			Placeholder: &slack.TextBlockObject{
				Type:     slack.PlainTextType,
				Text:     "Why is this device being excluded?",
				Emoji:    false,
				Verbatim: false,
			},
		},
		Hint: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "ex: top-secret",
			Emoji:    false,
			Verbatim: false,
		},
	}

	serial := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: SerialNumber,
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Serial Number",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.PlainTextInputBlockElement{
			Type:     slack.METPlainTextInput,
			ActionID: SerialInput,
			Placeholder: &slack.TextBlockObject{
				Type:     slack.PlainTextType,
				Text:     "Serial Number",
				Emoji:    false,
				Verbatim: false,
			},
		},
		Hint: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "ex: ABC123",
			Emoji:    false,
			Verbatim: false,
		},
		Optional:       false,
		DispatchAction: false,
	}

	today := time.Now().Format("2006-01-02")
	datePicker := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: ExclusionDatePicker,
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Until",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.DatePickerBlockElement{
			Type:        slack.METDatepicker,
			ActionID:    DatePicker,
			InitialDate: today,
		},
		Hint: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     fmt.Sprintf("ex: %s", today),
			Emoji:    false,
			Verbatim: false,
		},
		Optional:       false,
		DispatchAction: false,
	}

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			reason,
			serial,
			datePicker,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:       slack.VTModal,
		Title:      titleText,
		Close:      closeText,
		Submit:     submitText,
		Blocks:     blocks,
		CallbackID: AdminExclusionModal,
	}

	vr, err := b.bot.SlackClient().OpenView(triggerID, modalRequest)
	if err != nil {
		b.log.Err(err).Send()
	}
	b.log.Trace().Interface("view_response", vr).Msg("exclusion modal")
}

// exclusionRequested handles the request for an exclusion allowing authorized users to approve or deny
func (b *Bot) exclusionRequested(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.BlockActions[0]

	_, _, err := b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().Message.Timestamp)
	if err != nil {
		b.log.Err(err).Msg("deleting exclusion request message")
	}

	devices := b.tables.ExclusionSerials(ctx.Callback().User.ID)
	switch action.ActionID {
	case YesExclusion:
		b.log.Info().Msgf("%s wants to request an exclusion", ctx.Callback().User.ID)

		if len(devices) == 0 {
			_, _, err := b.bot.SlackClient().PostMessage(ctx.Callback().User.ID,
				slack.MsgOptionText("You have no devices to request an exclusion for", false))
			if err != nil {
				b.log.Err(err).Msg("posting exclusion denial")
			}
			return
		}

		b.exclusionRequest(devices, ctx.Callback().TriggerID)
	case YesAddExclusion:
		b.log.Debug().Msgf("%s wants to add an exclusion", ctx.Callback().User.ID)

		b.exclusionAdd(ctx.Callback().TriggerID)
	}
}

// exclusionSubmit handles the submission of the exclusion request by the admin
// and extracts the values from the modal.
func (b *Bot) exclusionSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	dv := values[ExclusionDatePicker][DatePicker].SelectedDate
	serial := values["serial_number"]["serial_input"].Value
	reason := values[ExclusionReason][ExclusionInput].Value

	b.log.Debug().
		Str("user_id", ctx.Callback().User.ID).
		Str("serial", serial).
		Str("until", dv).
		Str("reason", reason).
		Msg("exlusion request")

	ts, err := time.Parse("2006-01-02", dv)
	if err != nil {
		b.log.Debug().
			Str("date", dv).
			Str("serial", serial).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	// first we need to check if the serial exists
	err = b.tables.AddExclusion(serial, reason, ts)
	if err != nil {
		b.log.Err(err).Str("adding exclusion", "failed").Send()
	}

	_, _, err = b.bot.SlackClient().PostMessage(ctx.Callback().User.ID,
		slack.MsgOptionText("exclusion has been submitted :white_check_mark:", false))
	if err != nil {
		b.log.Err(err).Msg("posting ack for exclusion request")
	}
}
