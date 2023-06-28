package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// addExclusion adds a device to be excluded. Only authorized users can add exclusions.
func (b *Bot) addExclusion() *slacker.CommandDefinition {
	return &slacker.CommandDefinition{
		Description: "Add a device to be excluded",
		Examples:    []string{"add exclusion"},
		AuthorizationFunc: func(botCtx slacker.BotContext, request slacker.Request) bool {
			return helpers.Contains(b.cfg.authUsers, botCtx.Event().User)
		},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			b.wantAddExclusion(botCtx.Event().User, "Would you like add a device to the exclusion list?")
		},
		HideHelp: false,
	}
}

// exclusionApprove allows those in the approver channel to approve or deny the exclusion.
func (b *Bot) exclusionApprove(user, reason, until string, serials []string) {
	attachment := slack.Attachment{
		Title:      fmt.Sprintf("<@%s> is requesting an exclusion.", user),
		CallbackID: "exclusion_approver",
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
				Name:  "approve_exclusion",
				Text:  "Approve",
				Type:  "button",
				Value: "approve_exclusion",
				Style: "primary",
			},
			{
				Name:  "deny_exclusion",
				Text:  "Deny",
				Type:  "button",
				Value: "deny_exclusion",
				Style: "danger",
			},
		},
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := b.bot.Client().PostMessage(b.cfg.SlackAlertChannel, message)
	if err != nil {
		b.log.Err(err).Msg("posting message")
	}

	b.log.Debug().
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Msg("exclusion approval message sent")
}

// exclusionRequest is the modal presented to the admins who requested to add
// an exclusion to the exclusion list to load the exclusion into the database.
func (b *Bot) exclusionRequest(triggerID string) {
	headerText := slack.NewTextBlockObject(slack.MarkdownType, "Add Exclusion", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	reason := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: "exclusion_reason",
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Exclusion Reason",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.PlainTextInputBlockElement{
			Type:     slack.METPlainTextInput,
			ActionID: "exclusion_input",
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
		BlockID: "serial_number",
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Serial Number",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.PlainTextInputBlockElement{
			Type:     slack.METPlainTextInput,
			ActionID: "serial_input",
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
		BlockID: "exclusion_date_picker",
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Until",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.DatePickerBlockElement{
			Type:        slack.METDatepicker,
			ActionID:    "datePicker",
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
		CallbackID: "admin_exclusion_modal",
	}

	vr, err := b.bot.Client().OpenView(triggerID, modalRequest)
	if err != nil {
		b.log.Err(err).Send()
	}
	b.log.Trace().Interface("view_response", vr).Msg("exclusion modal")
}

// exclusionSubmit handles the submission of the exclusion request by the admin
// and extracts the values from the modal.
func (b *Bot) exclusionSubmit(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	values := callback.View.State.Values
	dv := values["exclusion_date_picker"]["datePicker"].SelectedDate
	serial := values["serial_number"]["serial_input"].Value
	reason := values["exclusion_reason"]["exclusion_input"].Value

	b.log.Debug().
		Str("user_id", callback.User.ID).
		Str("serial", serial).
		Str("until", dv).
		Str("reason", reason).
		Msg("exlusion request")

	s.SocketMode().Ack(*event.Request)

	ts, err := time.Parse("2006-01-02", dv)
	if err != nil {
		b.log.Debug().
			Str("date", dv).
			Str("serial", serial).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	// first we need to check if the serial exists
	err = b.db.AddExclusion(serial, reason, ts)
	if err != nil {
		b.log.Err(err).Str("adding exclusion", "failed").Send()
	}

	_, _, err = s.Client().PostMessage(callback.User.ID,
		slack.MsgOptionText("Exclusion has been submitted :white_check_mark:", false))
	if err != nil {
		b.log.Err(err).Msg("posting ack for exclusion request")
	}
}

// wantAddExclusion asks the user if they want to add an exclusion.
// this sends the user the modal created in exclusionRequest.
func (b *Bot) wantAddExclusion(user, text string) {
	b.modalGateway(
		&modalGateway{
			text:       text,
			callbackID: "exclusion_add_question",
			yesName:    "yes_add_exclusion",
			yesText:    "Yes",
			yesValue:   "yes_add_exclusion",
			yesStyle:   "primary",
			noName:     "no_exclusion",
			noText:     "No",
			noValue:    "no_exclusion",
			noStyle:    "danger",
			channel:    user,
			msg:        "exclusion request sent",
		},
	)
}
