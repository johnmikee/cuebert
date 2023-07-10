package bot

import (
	"fmt"
	"time"

	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

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
		BlockID: ReminderPickerHeader,
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	datePicker := slack.InputBlock{
		Type:    slack.MBTInput,
		BlockID: DatePicker,
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Date",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.DatePickerBlockElement{
			Type:        slack.METDatepicker,
			ActionID:    DatePicker,
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
		BlockID: TimePicker,
		Label: &slack.TextBlockObject{
			Type:     slack.PlainTextType,
			Text:     "Time",
			Emoji:    false,
			Verbatim: false,
		},
		Element: &slack.TimePickerBlockElement{
			Type:     slack.METTimepicker,
			ActionID: TimePicker,
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
		CallbackID:    ReminderPicker,
		ClearOnClose:  false,
		NotifyOnClose: false,
	}

	vr, err := b.bot.SlackClient().OpenView(triggerID, modalRequest)
	if err != nil {
		b.log.Err(err).Send()
	}
	b.log.Trace().Interface("vr", vr).Msg("reminder picker modal opened")
}

// reminderRequested is the callback for the reminder button
func (b *Bot) reminderRequested(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.BlockActions[0]

	if action.Value == RemindMe {
		b.log.Info().Msgf("%s wants a reminder to update", ctx.Callback().User.ID)

		err := b.tables.ACKACKD(ctx.Callback().User.ID, time.Now().UTC())
		if err != nil {
			b.log.Err(err).Msg("could not record the first ack time")
		}
	}

	_, _, err := b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().Message.Timestamp)
	if err != nil {
		b.log.Err(err).Msg("could not delete the reminder request message")
	}

	b.reminderPicker(ctx.Callback().TriggerID, "Please enter a time to be reminded")
}

// ScheduleReminder will execute the scheduled reminder set by the user.
func (b *Bot) ScheduleReminder(t time.Duration, ri *ReminderInfo) {
	sent, err := b.tables.ReminderSentCheck(ri.User)

	if err != nil {
		b.log.Info().
			Str("user", ri.User).
			AnErr("checking if reminder has been sent", err).
			Send()
		return
	}

	if sent {
		b.log.Debug().
			Str("user", ri.User).
			Bool("sent", sent).
			Msg("not sending reminder, already sent")
		return
	}

	b.log.Debug().
		Str("user", ri.User).
		Float64("sleeping_seconds", t.Seconds()).
		Msg("waiting to send reminder")

	time.Sleep(t)

	b.log.Trace().
		Str("user", ri.User).
		Str("serial", ri.Serial).
		Msg("sending requested reminder")

	err = b.deliverReminder(ri)

	if err != nil {
		b.log.Info().
			Str("user", ri.User).
			AnErr("delivering reminder", err).
			Send()
		return
	}

	err = b.tables.ReminderSent(true, ri.Serial)

	if err != nil {
		b.log.Info().
			Str("user", ri.User).
			Str("serial", ri.Serial).
			Str("table", "bot_results").
			AnErr("updating table", err).
			Send()
		return
	}

	b.log.Trace().
		Str("user", ri.User).
		Str("serial", ri.Serial).
		Str("table", "bot_results").
		Bool("updated", true).
		Msg("notification reminder sent")
}
