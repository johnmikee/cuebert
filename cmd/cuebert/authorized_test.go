package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
)

/*
The vast majority of the functions that are bot actions are really
difficult to test. Short of spinning up a Slack instance and running
the bot, there's not a lot of ways to _really_ test them.

What we can test is the output of the messages that are sent to Slack.
These are just json, so we can compare them to what we expect them to be.
*/
func TestBot_addExclusion(t *testing.T) {
	b := &Bot{
		cfg: &CuebertConfig{
			authUsers: []string{"user1"},
		},
	}
	ex := b.addExclusion()

	assert.Equal(t, "Add a device to be excluded", ex.Description)
	assert.Equal(t, []string{"add exclusion"}, ex.Examples)
	assert.False(t, ex.HideHelp)
	assert.NotNil(t, ex.AuthorizationFunc)
}

func TestBot_exclusionApprove(t *testing.T) {
	attachment := slack.Attachment{
		Title:      fmt.Sprintf("<@%s> is requesting an exclusion.", "user"),
		CallbackID: "exclusion_approver",
		Color:      "#3AA3E3",
		Fields: []slack.AttachmentField{
			{
				Title: "Serial Numbers",
				Value: strings.Join([]string{"123"}, ", "),
			},
			{
				Title: "Reason",
				Value: "reason",
			},
			{
				Title: "Until",
				Value: "until",
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

	assert.Equal(t, helpers.RespToJson(attachment), `{"color":"#3AA3E3","callback_id":"exclusion_approver","title":"\u003c@user\u003e is requesting an exclusion.","fields":[{"title":"Serial Numbers","value":"123","short":false},{"title":"Reason","value":"reason","short":false},{"title":"Until","value":"until","short":false}],"actions":[{"name":"approve_exclusion","text":"Approve","style":"primary","type":"button","value":"approve_exclusion"},{"name":"deny_exclusion","text":"Deny","style":"danger","type":"button","value":"deny_exclusion"}],"blocks":null}`)
}

func TestBot_exclusionRequest(t *testing.T) {
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

	fmt.Println(modalRequest)
}
