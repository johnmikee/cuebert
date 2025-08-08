package ui

import (
	"time"

	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// Example of refactored UserUpdate function using the UI Builder
// This replaces the 85-line monolithic function with a clean, maintainable approach

// UserUpdateModalHandler handles user update modal creation.
type UserUpdateModalHandler struct {
	builder *Builder
}

// NewUserUpdateModalHandler creates a new handler.
func NewUserUpdateModalHandler() *UserUpdateModalHandler {
	return &UserUpdateModalHandler{
		builder: New(),
	}
}

// CreateModal creates the user update modal
// Before: 85 lines of UI construction code mixed with business logic
// After: Clean separation of concerns.
func (h *UserUpdateModalHandler) CreateModal(user UserInfo) slack.ModalViewRequest {
	// Clear and configure the builder
	h.builder.Clear().
		SetTitle("Update User").
		SetClose("Cancel").
		SetSubmit("Update").
		AddHeader("Update User Information")

	// Add user fields
	h.addUserFields(user)
	h.addMessageTrackingFields(user)
	h.addTimezoneField(user)

	// Build the modal with metadata
	return h.builder.BuildModal(
		"user_update_modal",
		"",
		user.SlackID, // Store user ID in private metadata
	)
}

// addUserFields adds basic user information fields.
func (h *UserUpdateModalHandler) addUserFields(user UserInfo) {
	h.builder.
		AddEmailInput("email", "Email", user.Email, user.Email, true).
		AddTextInput(TextInput{
			Label:       "Full Name",
			Placeholder: user.FullName,
			BlockID:     "full_name",
			ActionID:    "full_name",
			Initial:     user.FullName,
			Optional:    true,
		})
}

// addMessageTrackingFields adds message tracking fields.
func (h *UserUpdateModalHandler) addMessageTrackingFields(user UserInfo) {
	// First message acknowledgment
	h.builder.AddYesNoRadio(
		"first_ack",
		"First Message Acknowledged",
		"first_ack_option",
		true,
	)

	// First message acknowledgment time
	if !user.FirstACKTime.IsZero() {
		h.builder.AddDatePicker(DatePickerInput{
			Label:    "First Message Acknowledged At",
			BlockID:  "first_ack_at",
			ActionID: "first_ack_date",
			Initial:  &user.FirstACKTime,
			Optional: true,
		})
	}

	// First message sent
	h.builder.AddYesNoRadio(
		"first_sent",
		"First Message Sent",
		"first_sent_option",
		true,
	)

	// First message sent time
	if !user.FirstMessageSentAt.IsZero() {
		h.builder.AddDatePicker(DatePickerInput{
			Label:    "First Message Sent At",
			BlockID:  "first_sent_at",
			ActionID: "first_sent_date",
			Initial:  &user.FirstMessageSentAt,
			Optional: true,
		})
	}
}

// addTimezoneField adds timezone configuration.
func (h *UserUpdateModalHandler) addTimezoneField(user UserInfo) {
	h.builder.AddSelectMenu(
		"tz_offset",
		"Timezone",
		"timezone_select",
		"Select timezone",
		getTimezoneOptions(),
		true,
	)
}

// UserInfo represents user information (simplified).
type UserInfo struct {
	SlackID            string
	Email              string
	FullName           string
	FirstACK           bool
	FirstACKTime       time.Time
	FirstMessageSent   bool
	FirstMessageSentAt time.Time
	TZOffset           int
}

// getTimezoneOptions returns timezone options.
func getTimezoneOptions() []SelectOption {
	return []SelectOption{
		{Text: "UTC-8 (PST)", Value: "-8"},
		{Text: "UTC-7 (MST)", Value: "-7"},
		{Text: "UTC-6 (CST)", Value: "-6"},
		{Text: "UTC-5 (EST)", Value: "-5"},
		{Text: "UTC+0 (GMT)", Value: "0"},
		{Text: "UTC+1 (CET)", Value: "1"},
	}
}

// ProcessModalSubmission processes the modal submission
// This separates the business logic from UI construction.
func ProcessModalSubmission(ctx *slacker.InteractionContext) (map[string]interface{}, error) {
	values := ctx.Callback().View.State.Values

	// Extract values in a clean, structured way
	updates := make(map[string]interface{})

	// Email
	if email := values["email"]["email"].Value; email != "" {
		updates["email"] = email
	}

	// First acknowledgment
	if firstAck := values["first_ack"]["first_ack_option"].SelectedOption.Value; firstAck != "" {
		updates["first_ack"] = firstAck == "yes"
	}

	// First acknowledgment date
	if firstAckDate := values["first_ack_at"]["first_ack_date"].SelectedDate; firstAckDate != "" {
		updates["first_ack_date"] = firstAckDate
	}

	// Timezone
	if tz := values["tz_offset"]["timezone_select"].SelectedOption.Value; tz != "" {
		updates["tz_offset"] = tz
	}

	return updates, nil
}

// Example of creating different modal types with the builder

// CreateTaskSelectionModal creates a task selection modal.
func CreateTaskSelectionModal(tasks []string) slack.ModalViewRequest {
	builder := New().
		SetTitle("Select Task").
		SetClose("Cancel").
		SetSubmit("Select").
		AddHeader("Choose a task to work on").
		AddSection("Select from the available tasks below:")

	// Convert tasks to options
	var options []RadioOption
	for _, task := range tasks {
		options = append(options, RadioOption{
			Text:  task,
			Value: task,
		})
	}

	builder.AddRadioButtons("task", "Available Tasks", "task_select", options, false)

	return builder.BuildModal("task_selection", "", "")
}

// CreateConfirmationModal creates a confirmation modal.
func CreateConfirmationModal(action, details string) slack.ModalViewRequest {
	return New().
		SetTitle("Confirm Action").
		SetClose("Cancel").
		SetSubmit("Confirm").
		AddHeader("⚠️ Confirmation Required").
		AddSection("You are about to "+action).
		AddDivider().
		AddContext(details).
		AddCheckboxes("confirm", "Please confirm", "confirm_check", []SelectOption{
			{Text: "I understand the consequences of this action", Value: "confirmed"},
		}, false).
		BuildModal("confirmation_modal", "", action)
}

// CreateSettingsModal creates a settings configuration modal.
func CreateSettingsModal(currentSettings Settings) slack.ModalViewRequest {
	builder := New().
		SetTitle("Settings").
		SetClose("Cancel").
		SetSubmit("Save")

	// Notification settings
	builder.
		AddHeader("📬 Notification Preferences").
		AddRadioButtons("notify_first", "First Message", "first_msg_notify", []RadioOption{
			{Text: "Enabled", Value: "true", Desc: "Receive first message notifications"},
			{Text: "Disabled", Value: "false", Desc: "No first message notifications"},
		}, false).
		AddRadioButtons("notify_final", "Final Message", "final_msg_notify", []RadioOption{
			{Text: "Enabled", Value: "true", Desc: "Receive final message notifications"},
			{Text: "Disabled", Value: "false", Desc: "No final message notifications"},
		}, false)

	// Reminder settings
	builder.
		AddDivider().
		AddHeader("⏰ Reminder Settings").
		AddTextInput(TextInput{
			Label:       "Default Reminder Interval (minutes)",
			Placeholder: "60",
			BlockID:     "reminder_interval",
			ActionID:    "interval",
			Initial:     currentSettings.ReminderInterval,
			Optional:    false,
		})

	// Advanced settings
	builder.
		AddDivider().
		AddHeader("⚙️ Advanced Settings").
		AddCheckboxes("advanced", "Options", "advanced_options", []SelectOption{
			{Text: "Enable debug mode", Value: "debug"},
			{Text: "Send daily reports", Value: "daily_reports"},
			{Text: "Include manager in notifications", Value: "include_manager"},
		}, true)

	return builder.BuildModal("settings_modal", "", "")
}

// Settings represents application settings.
type Settings struct {
	NotifyFirst      bool
	NotifyFinal      bool
	ReminderInterval string
	Debug            bool
	DailyReports     bool
	IncludeManager   bool
}
