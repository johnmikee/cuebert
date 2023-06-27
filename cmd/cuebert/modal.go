package main

import (
	"fmt"
	"time"

	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type modalGateway struct {
	text       string
	fallback   string
	callbackID string
	yesName    string
	yesText    string
	yesValue   string
	yesStyle   string
	noName     string
	noText     string
	noValue    string
	noStyle    string
	channel    string
	msg        string
}

var (
	titleText  = slack.NewTextBlockObject(slack.PlainTextType, "Cuebert", false, false)
	closeText  = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	submitText = slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
)

// modalGateway is a helper function to send a request to invoke a modal with two buttons.
func (b *Bot) modalGateway(m *modalGateway) {
	attachment := slack.Attachment{
		Text:       m.text,
		Fallback:   m.fallback,
		CallbackID: m.callbackID,
		Color:      "#3AA3E3",
		Actions: []slack.AttachmentAction{
			{
				Name:  m.yesName,
				Text:  m.yesText,
				Type:  "button",
				Value: m.yesValue,
				Style: m.yesStyle,
			},
			{
				Name:  m.noName,
				Text:  m.noText,
				Type:  "button",
				Value: m.noValue,
				Style: m.noStyle,
			},
		},
	}

	b.sendAttachment(attachment, m.channel, m.msg)
}

// modalSubmit is the handler for all modal submissions
func (b *Bot) modalSubmit(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	switch callback.View.CallbackID {
	case "admin_exclusion_modal":
		b.exclusionSubmit(s, event, callback)
	case "exception_modal":
		b.exceptionSubmit(s, event, callback)
	case "reminder_picker":
		b.reminderSubmit(s, event, callback)
	case START, RELOAD:
		b.loadInput(s, event, callback, callback.View.CallbackID)
	default:
		b.log.Debug().Msg("not something we are handling.")
		return
	}
}

// exceptionRequest is the modal the user will see when they request an exception
// the results of this modal will be sent to exceptionRequestDecision
func (b *Bot) exceptionRequest(devices []string, triggerID string) {
	headerSection := slack.SectionBlock{
		Type: slack.MBTSection,
		Text: &slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: "Request an exception",
		},
		BlockID: "exception_reason_header",
	}

	reason := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: "exception_reason",
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Exception Reason",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.PlainTextInputBlockElement{
			Type:     slack.METPlainTextInput,
			ActionID: "exception_input",
			Placeholder: &slack.TextBlockObject{
				Type:     slack.PlainTextType,
				Text:     "ex: I left my computer on the moon",
				Emoji:    false,
				Verbatim: false,
			},
		},
		Hint: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Why do you need an exception?",
			Emoji:    false,
			Verbatim: false,
		},
		Optional:       false,
		DispatchAction: false,
	}

	checkboxBlock := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: "user_devices",
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Which Device?",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.CheckboxGroupsBlockElement{
			Type:     slack.METCheckboxGroups,
			ActionID: "device_box",
			Options:  createOptionBlockObjects(devices),
		},
		Hint:           nil,
		Optional:       false,
		DispatchAction: false,
	}

	today := time.Now().Format("2006-01-02")
	datePicker := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: "exception_date_picker",
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Date",
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
			checkboxBlock,
			reason,
			datePicker,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:          slack.ViewType(slack.VTModal),
		Title:         titleText,
		Blocks:        blocks,
		Close:         closeText,
		Submit:        submitText,
		CallbackID:    "exception_modal",
		ClearOnClose:  false,
		NotifyOnClose: false,
	}

	vr, err := b.bot.Client().OpenView(triggerID, modalRequest)
	if err != nil {
		b.log.Err(err).Send()
	}

	b.log.Trace().Interface("vr", vr).Msg("exception request modal opened")
}

// reminderPicker is the modal the user will see when they request a reminder.
//
// does a quick check to make sure the date selected is not in the past. if it is,
// the modal will be re-opened letting the user know to pick a date in the future.
func (b *Bot) reminderPicker(triggerID, title string) {
	headerSection := slack.SectionBlock{
		Type: slack.MBTSection,
		Text: &slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: title,
		},
		BlockID: "reminder_picker_header",
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	datePicker := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: "date_picker",
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Date",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.DatePickerBlockElement{
			Type:        slack.METDatepicker,
			ActionID:    "datePicker",
			InitialDate: yesterday,
		},
		Hint: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     fmt.Sprintf("ex: %s", yesterday),
			Emoji:    false,
			Verbatim: false,
		},
		Optional:       false,
		DispatchAction: false,
	}

	timePicker := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: "time_picker",
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Time",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.TimePickerBlockElement{
			Type:     slack.METTimepicker,
			ActionID: "timePicker",
		},
		Hint: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "ex: 1:37 PM",
			Emoji:    false,
			Verbatim: false,
		},
		Optional:       false,
		DispatchAction: false,
	}

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			datePicker,
			timePicker,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:          slack.ViewType(slack.VTModal),
		Title:         titleText,
		Blocks:        blocks,
		Close:         closeText,
		Submit:        submitText,
		CallbackID:    "reminder_picker",
		ClearOnClose:  false,
		NotifyOnClose: false,
		ExternalID:    "",
	}

	vr, err := b.bot.Client().OpenView(triggerID, modalRequest)
	if err != nil {
		b.log.Err(err).Send()
	}
	b.log.Trace().Interface("vr", vr).Msg("reminder picker modal opened")
}

// createOptionBlockObjects - utility function for generating option block objects
func createOptionBlockObjects(options []string) []*slack.OptionBlockObject {
	optionBlockObjects := make([]*slack.OptionBlockObject, 0, len(options))

	for _, o := range options {
		optionText := slack.NewTextBlockObject(slack.PlainTextType, o, false, false)
		optionBlockObjects = append(optionBlockObjects, slack.NewOptionBlockObject(o, optionText, nil))
	}

	return optionBlockObjects
}
