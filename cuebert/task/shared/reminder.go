package shared

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

type Reminder struct {
	Text     string
	Deadline string
	Cutoff   string
	User     string
	Task     string
	Title    string
	Fields   []slack.AttachmentField
}

type ReminderPayload struct {
	UserServiceID string
	UserSlackID   string
	UserName      string
	Task          string
	FirstMessage  string
	FinalMessage  string
	Item          string
	Text          string
	TZOffset      int64
	Title         string
}

// DeliverReminder delivers the reminder to the user
func (c *Shared) DeliverReminder(r *Reminder) error {
	attachment := slack.Attachment{
		Title:      r.Title,
		Text:       r.Text,
		CallbackID: c.constFormat(UserReminder),
		Color:      "#3AA3E3",
		Fields:     r.Fields,
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := c.bot.Bot().SlackClient().PostMessage(r.User, message)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("posting message")
		return err
	}

	c.log.Debug().
		Str("user", r.User).
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Bool("sent", true).
		Msg("delivered reminder")

	return nil
}

// ReminderPicker is the modal the user will see when they request a reminder.
//
// does a quick check to make sure the date selected is not in the past. if it is,
// the modal will be re-opened letting the user know to pick a date in the future.
func (c *Shared) ReminderPicker(triggerID, title string) {
	headerSection := slack.SectionBlock{
		Type: slack.MBTSection,
		Text: &slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: title,
		},
		BlockID: ReminderPickerHeader,
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	datePickerSection := slack.NewTextBlockObject(slack.PlainTextType, "Date", false, false)
	datePickerHint := slack.NewTextBlockObject(slack.PlainTextType, fmt.Sprintf("ex: %s", yesterday), false, false)
	datePickerInput := slack.NewDatePickerBlockElement(DatePicker)
	datePicker := slack.NewInputBlock(DatePicker, datePickerSection, datePickerHint, datePickerInput)

	timePickSection := slack.NewTextBlockObject(slack.PlainTextType, "Time", false, false)
	timePickerHint := slack.NewTextBlockObject(slack.PlainTextType, "ex: 1:37 PM", false, false)
	timePickerInput := slack.NewTimePickerBlockElement(TimePicker)
	timePicker := slack.NewInputBlock(TimePicker, timePickSection, timePickerHint, timePickerInput)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			datePicker,
			timePicker,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:          slack.ViewType(slack.VTModal),
		Title:         TitleText,
		Blocks:        blocks,
		Close:         CloseText,
		Submit:        SubmitText,
		CallbackID:    c.constFormat(ReminderPicker),
		ClearOnClose:  false,
		NotifyOnClose: false,
	}

	_, err := c.bot.Bot().SlackClient().OpenView(triggerID, modalRequest)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("opening modal")
	}
}

// ReminderRequested is the callback for the reminder button
func (c *Shared) ReminderRequested(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0]

	_, _, err := c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not delete the reminder request message")
	}

	if action.Name == RemindMe {
		c.log.Info().Str("user", ctx.Callback().User.ID).Msg("wants reminder to update")

		err := c.caller.MessageAcknowledged(ctx.Callback().User.ID, time.Now().UTC(), "reminder")
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not record the time of the reminder request")
		}
		c.log.Info().Str("user", ctx.Callback().User.ID).Msg("opening reminder picker")
		c.ReminderPicker(ctx.Callback().TriggerID, "Please enter a time to be reminded")
	}
}

// ReminderSubmit is the callback for the reminder picker
func (c *Shared) ReminderSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values

	dv := values[DatePicker][DatePicker]
	tv := values[TimePicker][TimePicker]

	c.log.Trace().
		Str("user", ctx.Callback().User.ID).
		Str("date", dv.SelectedDate).Str("time", dv.SelectedTime).
		Msg("is setting a reminder for updating")

	offset, err := c.caller.GetUserTZ(ctx.Callback().User.ID)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not get user")
		return
	}
	c.log.Info().Int64("offset", offset).Msg("user offset")
	// validate this is not in the past.
	if !helpers.FutureDate(dv.SelectedDate, tv.SelectedTime, offset) {
		c.log.Debug().Str("user", ctx.Callback().User.ID).Msg("set a date in the past.")

		update := fmt.Sprintf(
			"Sorry %s %s already happened..\nPlease set a date in the future. :clock1:",
			dv.SelectedDate,
			tv.SelectedTime,
		)
		_, _, err := c.bot.Bot().SlackClient().PostMessage(ctx.Callback().User.ID, slack.MsgOptionText(update, false))
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("posting time fix message")
		}
		return
	}
	c.log.Info().Str("user", ctx.Callback().User.ID).Msg("setting reminder")

	var update string
	err = c.caller.UpdateReminderTime(dv.SelectedDate, tv.SelectedTime, ctx.Callback().User.ID)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("updating reminder time")
		update = "Oops something went wrong! Please try again."
	} else {
		update = fmt.Sprintf("Your reminder has been set for %s %s :clock1:", dv.SelectedDate, tv.SelectedTime)
	}

	_, _, err = c.bot.Bot().SlackClient().PostMessage(ctx.Callback().User.ID, slack.MsgOptionText(update, false))
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("posting time fix message")
	}
}
