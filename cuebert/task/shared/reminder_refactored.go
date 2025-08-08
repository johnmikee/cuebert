package shared

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/bot/ui"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// ReminderRefactored contains refactored reminder data
type ReminderRefactored struct {
	Text     string
	Deadline string
	Cutoff   string
	User     string
	Task     string
	Title    string
	Fields   []slack.AttachmentField
}

// DeliverReminderRefactored delivers the reminder to the user with improved error handling
func (c *Shared) DeliverReminderRefactored(r *ReminderRefactored) error {
	// Build the reminder message using UI builder for consistency
	builder := ui.New()

	// Create a rich message block instead of legacy attachments
	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", r.Title, false, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn", r.Text, false, false),
			nil,
			nil,
		),
	}

	// Add fields if present
	if len(r.Fields) > 0 {
		fields := make([]*slack.TextBlockObject, 0, len(r.Fields))
		for _, field := range r.Fields {
			fields = append(fields, slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("*%s*\n%s", field.Title, field.Value), false, false))
		}

		blocks = append(blocks, slack.NewSectionBlock(nil, fields, nil))
	}

	// Add action buttons if needed
	elements := []slack.BlockElement{
		slack.NewButtonBlockElement(
			"complete_task",
			"complete",
			slack.NewTextBlockObject("plain_text", "Mark Complete", false, false),
		).WithStyle(slack.StylePrimary),
		slack.NewButtonBlockElement(
			"snooze_reminder",
			"snooze",
			slack.NewTextBlockObject("plain_text", "Remind Later", false, false),
		),
	}

	blocks = append(blocks, slack.NewActionBlock(c.constFormat(UserReminder), elements...))

	// Send the message
	channelID, timestamp, err := c.bot.Bot().SlackClient().PostMessage(
		r.User,
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		c.log.Error().
			Err(errors.SlackError("post reminder message", err)).
			Str("user", r.User).
			Msg("failed to deliver reminder")
		return errors.SlackError("deliver reminder", err)
	}

	c.log.Debug().
		Str("user", r.User).
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Bool("sent", true).
		Msg("delivered reminder")

	return nil
}

// ReminderPickerRefactored creates a reminder picker modal using the UI builder
func (c *Shared) ReminderPickerRefactored(triggerID, title string) {
	// Get example date/time for placeholders
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	exampleTime := "2:00 PM"

	// Build the modal using the UI builder
	modalRequest := ui.New().
		SetTitle("Set Reminder").
		SetClose("Cancel").
		SetSubmit("Set Reminder").
		AddHeader(title).
		AddDatePicker(ui.DatePickerInput{
			Label:       "Reminder Date",
			BlockID:     DatePicker,
			ActionID:    DatePicker,
			Placeholder: fmt.Sprintf("Select date (e.g., %s)", tomorrow),
			Optional:    false,
		}).
		AddTimePicker("time_picker", "Reminder Time", "time_picker",
			fmt.Sprintf("Select time (e.g., %s)", exampleTime), false).
		AddContext("⏰ Times are in your local timezone").
		BuildModal(c.constFormat(ReminderPicker), "", "")

	// Open the modal
	if _, err := c.bot.Bot().SlackClient().OpenView(triggerID, modalRequest); err != nil {
		c.log.Error().
			Err(errors.SlackError("open reminder picker", err)).
			Str("trigger_id", triggerID).
			Msg("failed to open reminder picker modal")
	}
}

// ReminderRequestedRefactored handles reminder requests with improved error handling
func (c *Shared) ReminderRequestedRefactored(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0]
	userID := ctx.Callback().User.ID

	// Delete the original message
	if _, _, err := c.bot.Bot().SlackClient().DeleteMessage(
		ctx.Callback().Channel.ID,
		ctx.Callback().MessageTs,
	); err != nil {
		c.log.Info().
			Err(errors.SlackError("delete reminder request", err)).
			Str("user_id", userID).
			Msg("could not delete the reminder request message")
	}

	if action.Name != RemindMe {
		return
	}

	c.log.Info().Str("user", userID).Msg("user requested reminder")

	// Record the acknowledgment
	if err := c.caller.MessageAcknowledged(userID, time.Now().UTC(), "reminder"); err != nil {
		c.log.Error().
			Err(errors.DBError("record reminder acknowledgment", err)).
			Str("user_id", userID).
			Msg("could not record the time of the reminder request")
	}

	// Open the reminder picker
	c.log.Info().Str("user", userID).Msg("opening reminder picker")
	c.ReminderPickerRefactored(ctx.Callback().TriggerID, "📅 When should I remind you?")
}

// ReminderSubmitRefactored handles reminder submission with validation and error handling
func (c *Shared) ReminderSubmitRefactored(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	userID := ctx.Callback().User.ID

	// Extract date and time values
	selectedDate := values[DatePicker][DatePicker].SelectedDate
	selectedTime := values["time_picker"]["time_picker"].SelectedTime

	c.log.Trace().
		Str("user", userID).
		Str("date", selectedDate).
		Str("time", selectedTime).
		Msg("processing reminder request")

	// Get user timezone
	offset, err := c.caller.GetUserTZ(userID)
	if err != nil {
		c.log.Error().
			Err(errors.DBError("get user timezone", err)).
			Str("user_id", userID).
			Msg("could not get user timezone")

		// Send error message to user
		c.sendReminderError(userID, "Unable to set reminder: timezone not found")
		return
	}

	// Parse and validate the reminder time
	userLocation := helpers.GenLocation(offset)
	reminderTime, err := c.parseReminderTime(selectedDate, selectedTime, userLocation)
	if err != nil {
		c.log.Error().
			Err(errors.ValidationError("invalid reminder time")).
			Str("user_id", userID).
			Str("date", selectedDate).
			Str("time", selectedTime).
			Msg("invalid reminder time")

		// Reopen modal with error message
		c.ReminderPickerRefactored(ctx.Callback().TriggerID,
			"❌ Please select a future date and time")
		return
	}

	// Ensure reminder is in the future
	if reminderTime.Before(time.Now()) {
		c.log.Info().
			Str("user_id", userID).
			Time("selected_time", reminderTime).
			Msg("user selected past time for reminder")

		// Reopen modal with error message
		c.ReminderPickerRefactored(ctx.Callback().TriggerID,
			"❌ Please select a future date and time")
		return
	}

	// Set the reminder
	if err := c.caller.SetReminder(userID, reminderTime); err != nil {
		c.log.Error().
			Err(errors.DBError("set reminder", err)).
			Str("user_id", userID).
			Time("reminder_time", reminderTime).
			Msg("failed to set reminder")

		c.sendReminderError(userID, "Unable to set reminder. Please try again.")
		return
	}

	// Send confirmation
	c.sendReminderConfirmation(userID, reminderTime, userLocation)
}

// parseReminderTime parses the date and time strings into a time.Time
func (c *Shared) parseReminderTime(dateStr, timeStr string, loc *time.Location) (time.Time, error) {
	// Combine date and time strings
	dateTimeStr := fmt.Sprintf("%s %s", dateStr, timeStr)

	// Parse with the user's timezone
	reminderTime, err := time.ParseInLocation("2006-01-02 15:04", dateTimeStr, loc)
	if err != nil {
		return time.Time{}, err
	}

	return reminderTime, nil
}

// sendReminderError sends an error message to the user
func (c *Shared) sendReminderError(userID, errorMsg string) {
	blocks := []slack.Block{
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("❌ %s", errorMsg), false, false),
			nil,
			nil,
		),
	}

	if _, _, err := c.bot.Bot().SlackClient().PostMessage(
		userID,
		slack.MsgOptionBlocks(blocks...),
	); err != nil {
		c.log.Error().
			Err(errors.SlackError("send reminder error", err)).
			Str("user_id", userID).
			Msg("failed to send error message")
	}
}

// sendReminderConfirmation sends a confirmation message to the user
func (c *Shared) sendReminderConfirmation(userID string, reminderTime time.Time, loc *time.Location) {
	// Format the reminder time in user's timezone
	formattedTime := reminderTime.In(loc).Format("Monday, January 2, 2006 at 3:04 PM MST")

	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "✅ Reminder Set!", false, false),
		),
		slack.NewSectionBlock(
			slack.NewTextBlockObject("mrkdwn",
				fmt.Sprintf("I'll remind you on:\n*%s*", formattedTime), false, false),
			nil,
			nil,
		),
		slack.NewContextBlock("",
			slack.NewTextBlockObject("mrkdwn",
				"💡 _You can set another reminder when this one is sent_", false, false),
		),
	}

	if _, _, err := c.bot.Bot().SlackClient().PostMessage(
		userID,
		slack.MsgOptionBlocks(blocks...),
	); err != nil {
		c.log.Error().
			Err(errors.SlackError("send reminder confirmation", err)).
			Str("user_id", userID).
			Msg("failed to send confirmation message")
	}
}

// CreateReminderSettingsModal creates a modal for managing reminder preferences
func (c *Shared) CreateReminderSettingsModal(userID string) slack.ModalViewRequest {
	return ui.New().
		SetTitle("Reminder Settings").
		SetClose("Cancel").
		SetSubmit("Save").
		AddHeader("🔔 Reminder Preferences").
		AddRadioButtons("reminder_frequency", "Default Reminder Frequency", "frequency_select",
			[]ui.RadioOption{
				{Text: "Daily", Value: "daily", Desc: "Receive daily reminders"},
				{Text: "Every 3 days", Value: "3days", Desc: "Receive reminders every 3 days"},
				{Text: "Weekly", Value: "weekly", Desc: "Receive weekly reminders"},
				{Text: "Custom", Value: "custom", Desc: "Set a custom reminder schedule"},
			},
			false,
		).
		AddDivider().
		AddHeader("⏰ Preferred Reminder Time").
		AddTimePicker("preferred_time", "Default Time", "preferred_time_select",
			"9:00 AM", false).
		AddDivider().
		AddHeader("📧 Notification Settings").
		AddCheckboxes("notify_options", "Notification Channels", "notify_channels",
			[]ui.SelectOption{
				{Text: "Slack Direct Message", Value: "slack_dm"},
				{Text: "Email Notification", Value: "email"},
				{Text: "Mobile Push (if available)", Value: "mobile"},
			},
			true,
		).
		BuildModal("reminder_settings", "", userID)
}
