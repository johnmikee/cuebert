package shared

import (
	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/johnmikee/cuebert/bot/ui"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// UserSelectorRefactored demonstrates the refactored version using UI builder
func (c *Shared) UserSelectorRefactored(triggerID, channel string) {
	// Use the UI builder for clean modal construction
	modalRequest := ui.New().
		SetTitle("Cuebert").
		SetClose("Close").
		SetSubmit("Submit").
		AddHeader("Select User").
		AddUserSelect("update_user", "User", "users_select", nil, false).
		BuildModal(
			c.constFormat(UpdateUserSelector),
			channel,   // External ID
			triggerID, // Private metadata
		)

	vr, err := c.bot.Bot().SlackClient().OpenView(triggerID, modalRequest)
	if err != nil {
		c.log.Error().Err(errors.SlackError("open user selector modal", err)).Msg("error opening modal")
		return
	}

	c.log.Trace().Interface("view_response", vr).Msg("modal opened")
}

// ComplianceCheckerRefactored demonstrates improved error handling and cleaner code
func (c *Shared) ComplianceCheckerRefactored(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0]

	// Delete the original message
	if _, _, err := c.bot.Bot().SlackClient().DeleteMessage(
		ctx.Callback().Channel.ID,
		ctx.Callback().MessageTs,
	); err != nil {
		c.log.Info().Err(errors.SlackError("delete compliance message", err)).
			Msg("could not delete the compliance request message")
	}

	if action.Name != YesCheck {
		return
	}

	// Check compliance status
	userID := ctx.Callback().User.ID
	compliant, err := c.impl.CheckComplianceStatus(userID)

	// Determine the appropriate message
	message := c.determineComplianceMessage(compliant, err)

	// Update compliance status if user is compliant
	if compliant && err == nil {
		if updateErr := c.caller.UserCompliant(userID); updateErr != nil {
			c.log.Error().Err(errors.DBError("update user compliance", updateErr)).
				Str("user_id", userID).
				Msg("could not update user compliance")
			message += " ⚠️ There was an error updating your compliance status."
		}
	}

	// Send the response message
	if _, _, err := c.bot.Bot().SlackClient().PostMessage(
		userID,
		slack.MsgOptionText(message, false),
	); err != nil {
		c.log.Error().Err(errors.SlackError("send compliance result", err)).
			Str("user_id", userID).
			Msg("posting message")
	}
}

// determineComplianceMessage determines the appropriate message based on compliance status
func (c *Shared) determineComplianceMessage(compliant bool, err error) string {
	if err != nil {
		c.log.Info().Err(err).Bool("compliant", compliant).Msg("compliance check error")

		switch err.Error() {
		case "no user found", "no_text":
			return "✅"
		default:
			return "⚠️ Unable to verify compliance status. Please try again later."
		}
	}

	if compliant {
		c.log.Info().Msg("user is now compliant")
		return "You are now compliant! ✅"
	}

	// User is not compliant - determine the appropriate message
	if c.flags.TaskImplementation == string(implementation.OS) {
		return "⚠️ You have not completed the required update. Please update your device and try again."
	}

	return "⚠️ You have not completed the required training. Please complete the training and try again."
}

// WantUpdateUserRefactored demonstrates using the UI builder for simple modals
func (c *Shared) WantUpdateUserRefactored(ch string) {
	// Create a confirmation modal using the UI builder
	modal := ui.New().
		SetTitle("Update User").
		SetClose("Cancel").
		AddSection("Do you want to update a user's values?").
		AddRadioButtons("update_choice", "Choose an option", "update_user_choice",
			[]ui.RadioOption{
				{Text: "Yes", Value: UpdateUserYes, Desc: "Update user information"},
				{Text: "No", Value: UpdateUserNo, Desc: "Cancel this operation"},
			},
			false,
		).
		BuildMessage()

	// Send the interactive message
	if _, _, err := c.bot.Bot().SlackClient().PostMessage(
		ch,
		slack.MsgOptionBlocks(modal.Blocks.BlockSet...),
	); err != nil {
		c.log.Error().Err(errors.SlackError("send update user prompt", err)).
			Str("channel", ch).
			Msg("failed to send update user prompt")
	}
}

// UpdateUserInfoInteractiveRefactored demonstrates cleaner command registration
func (c *Shared) UpdateUserInfoInteractiveRefactored() {
	definition := &slacker.CommandDefinition{
		Command:     "update user interactive",
		Description: "Update info about a user",
		Examples:    []string{"update user interactive"},
		Middlewares: []slacker.CommandMiddlewareHandler{
			ops.AuthorizationMiddleware(c.flags.AuthUsers),
		},
		Handler: func(ctx *slacker.CommandContext) {
			// Add error handling and context
			defer func() {
				if r := recover(); r != nil {
					c.log.Error().
						Err(errors.SystemError("command panic", nil)).
						Interface("panic", r).
						Msg("update user interactive command panicked")

					if err := ctx.Response().Reply("An unexpected error occurred. Please try again."); err != nil {
						c.log.Error().Err(errors.SlackError("send panic message", err)).
							Msg("failed to send error message")
					}
				}
			}()

			c.WantUpdateUserRefactored(ctx.Event().ChannelID)
		},
	}

	c.bot.Bot().AddCommand(definition)
}

// CreateUserUpdateModal creates a comprehensive user update modal using the UI builder
func (c *Shared) CreateUserUpdateModal(user UserInfo) slack.ModalViewRequest {
	builder := ui.New().
		SetTitle("Update User Information").
		SetClose("Cancel").
		SetSubmit("Update")

	// Basic user information section
	builder.
		AddHeader("👤 User Information").
		AddEmailInput("email", "Email Address", user.Email, user.Email, false).
		AddTextInput(ui.TextInput{
			Label:       "Full Name",
			Placeholder: "Enter full name",
			BlockID:     "full_name",
			ActionID:    "full_name_input",
			Initial:     user.FullName,
			Optional:    false,
		}).
		AddTextInput(ui.TextInput{
			Label:       "Slack User ID",
			Placeholder: "U1234567890",
			BlockID:     "slack_id",
			ActionID:    "slack_id_input",
			Initial:     user.SlackID,
			Optional:    false,
		})

	// Compliance tracking section
	builder.
		AddDivider().
		AddHeader("📋 Compliance Tracking").
		AddYesNoRadio(
			"task_complete",
			"Task Completed",
			"task_complete_option",
			false,
		).
		AddDatePicker(ui.DatePickerInput{
			Label:    "Completion Date",
			BlockID:  "completion_date",
			ActionID: "completion_date_picker",
			Initial:  nil,
			Optional: true,
		})

	// Message tracking section
	builder.
		AddDivider().
		AddHeader("💬 Message Tracking").
		AddCheckboxes("message_status", "Message Status", "message_checkboxes",
			[]ui.SelectOption{
				{Text: "First message sent", Value: "first_sent"},
				{Text: "First message acknowledged", Value: "first_ack"},
				{Text: "Final message sent", Value: "final_sent"},
				{Text: "Reminder active", Value: "reminder_active"},
			},
			true,
		)

	// Timezone configuration
	builder.
		AddDivider().
		AddHeader("🌍 Timezone Settings").
		AddSelectMenu(
			"timezone",
			"User Timezone",
			"timezone_select",
			"Select timezone",
			[]ui.SelectOption{
				{Text: "Pacific Time (UTC-8)", Value: "-8"},
				{Text: "Mountain Time (UTC-7)", Value: "-7"},
				{Text: "Central Time (UTC-6)", Value: "-6"},
				{Text: "Eastern Time (UTC-5)", Value: "-5"},
				{Text: "UTC", Value: "0"},
				{Text: "Central European Time (UTC+1)", Value: "1"},
			},
			false,
		)

	// Build the modal with user ID in private metadata
	return builder.BuildModal("user_update_modal", "", user.SlackID)
}

// UserInfo represents user information for the modal
type UserInfo struct {
	SlackID      string
	Email        string
	FullName     string
	TaskComplete bool
	TZOffset     int
}
