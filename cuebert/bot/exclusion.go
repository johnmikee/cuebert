package bot

import (
	"fmt"
	"time"

	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// exclusionHelp invokes the exclusion modal users can use to request an exclusion
func (b *Bot) exclusionHelp() {
	definition := &slacker.CommandDefinition{
		Command:     "request exclusion",
		Description: "Request an exclusion",
		Examples:    []string{"request exclusion"},
		Handler: func(ctx *slacker.CommandContext) {
			_, err := ctx.Response().ReplyBlocks(
				[]slack.Block{
					slack.NewSectionBlock(
						slack.NewTextBlockObject("mrkdwn", "Would you like to request an exclusion for updating?", false, false), nil, nil),
					slack.NewActionBlock(
						ExclusionQuestion,
						slack.NewButtonBlockElement(
							YesExclusion,
							YesExclusion,
							slack.NewTextBlockObject("plain_text", Yes, false, false)).
							WithStyle(slack.StylePrimary),
						slack.NewButtonBlockElement(
							ExclusionNo,
							ExclusionNo,
							slack.NewTextBlockObject("plain_text", No, false, false)).
							WithStyle(slack.StyleDanger)),
				},
			)
			if err != nil {
				b.log.Debug().AnErr("sending exclusion prompt", err).
					Send()
			}
		},
	}
	b.bot.AddCommand(definition)
}

// exclusionRequest is the modal the user will see when they request an exclusion
// the results of this modal will be sent to exclusionRequestDecision
func (b *Bot) exclusionRequest(devices []string, triggerID string) {
	headerSection := slack.SectionBlock{
		Type: slack.MBTSection,
		Text: &slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: "Request an Exclusion",
		},
		BlockID: ExclusionReasonHeader,
	}

	reason := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: ExclusionReason,
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "exclusion Reason",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.PlainTextInputBlockElement{
			Type:     slack.METPlainTextInput,
			ActionID: ExclusionInput,
			Placeholder: &slack.TextBlockObject{
				Type:     slack.PlainTextType,
				Text:     "ex: I left my computer on the moon",
				Emoji:    false,
				Verbatim: false,
			},
		},
		Hint: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Why do you need an exclusion?",
			Emoji:    false,
			Verbatim: false,
		},
		Optional:       false,
		DispatchAction: false,
	}

	checkboxBlock := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: UserDevices,
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Which Device?",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.CheckboxGroupsBlockElement{
			Type:     slack.METCheckboxGroups,
			ActionID: DeviceBox,
			Options:  createOptionBlockObjects(devices),
		},
		Hint:           nil,
		Optional:       false,
		DispatchAction: false,
	}

	today := time.Now().Format("2006-01-02")
	datePicker := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: ExclusionDatePicker,
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Date",
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
		CallbackID:    ExclusionModal,
		ClearOnClose:  false,
		NotifyOnClose: false,
	}

	vr, err := b.bot.SlackClient().OpenView(triggerID, modalRequest)
	if err != nil {
		b.log.Err(err).Send()
	}

	b.log.Trace().Interface("vr", vr).Msg("exclusion request modal opened")
}

// after the user has selected the devices they want to exclude, this function grabs the values to send to the db.
func (b *Bot) userExclusionSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	dv := values[ExclusionDatePicker][DatePicker].SelectedDate
	reason := values[ExclusionReason][ExclusionInput].Value

	serials := exclusionSubmitSerials(ctx.Callback())
	user := ctx.Callback().User.ID
	b.log.Debug().
		Str("user", user).
		Str("date", dv).
		Str("reason", reason).
		Strs("serials", serials).
		Msg("exclusion request")

	ts, err := time.Parse("2006-01-02", dv)
	if err != nil {
		b.log.Debug().
			Str("date", dv).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	err = b.tables.RequestExclusion(ctx.Callback().User.ID, reason, serials, ts)
	if err != nil {
		b.log.Err(err).Msg("adding exclusion request to db")
	}

	_, _, err = b.bot.SlackClient().PostMessage(ctx.Callback().User.ID,
		slack.MsgOptionText("Your request has been submitted :white_check_mark:", false))
	if err != nil {
		b.log.Err(err).Msg("posting ack for exclusion request")
	}

	b.exclusionApprove(user, reason, dv, serials)
}

func exclusionSubmitSerials(callback *slack.InteractionCallback) []string {
	serials := []string{}
	for _, v := range callback.View.State.Values["user_devices"]["device_box"].SelectedOptions {
		serials = append(serials, v.Text.Text)
	}
	return serials
}
