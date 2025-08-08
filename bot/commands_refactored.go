package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/db/repository"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// CommandsRefactored demonstrates how to implement bot commands with improved error handling
// and repository pattern usage.
type CommandsRefactored struct {
	userRepo     repository.UserRepository
	trackingRepo repository.TrackingRepository
	log          logger.Logger
}

// HandleUserLookup demonstrates a refactored command handler with proper error handling.
func (c *CommandsRefactored) HandleUserLookup(ctx *slacker.CommandContext) {
	// Extract user parameter with validation
	userID := ctx.Request().StringParam("user", "")
	if userID == "" {
		if err := c.replyWithError(ctx, errors.ValidationError("user", "parameter is required")); err != nil {
			c.log.Error().Err(errors.SlackError("send validation error", err)).Msg("failed to send error message")
		}
		return
	}

	// Create context with timeout
	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Look up user using repository
	user, err := c.userRepo.FindBySlackID(reqCtx, userID)
	if err != nil {
		if errors.Is(err, errors.ErrCodeDBNotFound) {
			if err := c.replyWithError(ctx, errors.NotFoundError("user")); err != nil {
				c.log.Error().Err(errors.SlackError("send not found error", err)).Msg("failed to send error message")
			}
			return
		}

		c.log.Error().Err(errors.DBError("find user", err)).Str("user_id", userID).Msg("database error")
		if err := c.replyWithError(ctx, errors.Wrap(err, errors.ErrCodeInternal, "lookup user")); err != nil {
			c.log.Error().Err(errors.SlackError("send internal error", err)).Msg("failed to send error message")
		}
		return
	}

	// Format and send response
	response := c.formatUserResponse(user)
	_, err = ctx.Response().Reply(response)
	if err != nil {
		c.log.Error().
			Err(errors.SlackError("send user info", err)).
			Str("user_id", userID).
			Msg("failed to send user information")
	}
}

// HandleTaskStatus demonstrates error handling for task status checks.
func (c *CommandsRefactored) HandleTaskStatus(ctx *slacker.CommandContext) {
	// Get user from context
	event := ctx.Event()
	if event.UserID == "" {
		c.log.Error().Err(errors.ValidationError("user ID", "not present in event")).Msg("invalid event")
		return
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get user's tracking records
	trackingRecords, err := c.trackingRepo.FindByUserSlackID(reqCtx, event.UserID)
	if err != nil {
		c.log.Error().
			Err(errors.DBError("find tracking records", err)).
			Str("user_id", event.UserID).
			Msg("failed to get tracking records")

		if err := c.replyWithError(ctx, errors.Wrap(err, errors.ErrCodeInternal, "fetch task status")); err != nil {
			c.log.Error().Err(errors.SlackError("send error", err)).Msg("failed to send error message")
		}
		return
	}

	if len(trackingRecords) == 0 {
		_, err := ctx.Response().Reply("You have no active tasks.")
		if err != nil {
			c.log.Error().Err(errors.SlackError("send no tasks message", err)).Msg("failed to send message")
		}
		return
	}

	// Build status message
	blocks := c.buildTaskStatusBlocks(trackingRecords)
	_, err = ctx.Response().ReplyBlocks(blocks)
	if err != nil {
		c.log.Error().
			Err(errors.SlackError("send task status", err)).
			Str("user_id", event.UserID).
			Msg("failed to send task status")
	}
}

// HandleCompleteTask demonstrates error handling for task completion.
func (c *CommandsRefactored) HandleCompleteTask(ctx *slacker.CommandContext) {
	taskName := ctx.Request().StringParam("task", "")
	if taskName == "" {
		if err := c.replyWithError(ctx, errors.ValidationError("task", "parameter is required")); err != nil {
			c.log.Error().Err(errors.SlackError("send validation error", err)).Msg("failed to send error message")
		}
		return
	}

	event := ctx.Event()
	reqCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Mark task as complete
	err := c.trackingRepo.MarkTaskComplete(reqCtx, event.UserID, taskName)
	if err != nil {
		if errors.Is(err, errors.ErrCodeDBNotFound) {
			if err := c.replyWithError(ctx, errors.NotFoundError("task")); err != nil {
				c.log.Error().Err(errors.SlackError("send not found error", err)).Msg("failed to send error message")
			}
			return
		}

		c.log.Error().
			Err(errors.DBError("mark task complete", err)).
			Str("user_id", event.UserID).
			Str("task", taskName).
			Msg("failed to complete task")

		if err := c.replyWithError(ctx, errors.Wrap(err, errors.ErrCodeInternal, "complete task")); err != nil {
			c.log.Error().Err(errors.SlackError("send error", err)).Msg("failed to send error message")
		}
		return
	}

	// Send success message
	successMsg := fmt.Sprintf("✅ Task '%s' marked as complete!", taskName)
	_, err = ctx.Response().Reply(successMsg)
	if err != nil {
		c.log.Error().
			Err(errors.SlackError("send success message", err)).
			Msg("failed to send completion message")
	}

	// Log successful completion for audit
	c.log.Info().
		Str("user_id", event.UserID).
		Str("task", taskName).
		Msg("task completed successfully")
}

// replyWithError sends a user-friendly error message.
func (c *CommandsRefactored) replyWithError(ctx *slacker.CommandContext, err error) error {
	var message string

	// Check error type using error codes
	if appErr, ok := errors.As(err); ok {
		switch appErr.Code {
		case errors.ErrCodeNotFound, errors.ErrCodeDBNotFound:
			message = fmt.Sprintf("❌ %s", err.Error())
		case errors.ErrCodeValidation:
			message = fmt.Sprintf("⚠️ Invalid input: %s", err.Error())
		case errors.ErrCodePermission:
			message = "🚫 You don't have permission to perform this action."
		default:
			message = "❌ An error occurred. Please try again later or contact support."
		}
	} else {
		message = "❌ An error occurred. Please try again later or contact support."
	}

	_, replyErr := ctx.Response().Reply(message)
	return replyErr
}

// formatUserResponse formats user data for display.
func (c *CommandsRefactored) formatUserResponse(user *repository.User) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*User Information*\n"))
	sb.WriteString(fmt.Sprintf("• Name: %s\n", user.UserLongName))
	sb.WriteString(fmt.Sprintf("• Email: %s\n", user.UserEmail))
	sb.WriteString(fmt.Sprintf("• Slack ID: %s\n", user.UserSlackID))
	sb.WriteString(fmt.Sprintf("• Service ID: %s\n", user.UserServiceID))
	// Note: ManagerSlackID field may need to be added to the User model if manager tracking is needed
	return sb.String()
}

// buildTaskStatusBlocks creates Slack blocks for task status display.
func (c *CommandsRefactored) buildTaskStatusBlocks(records []repository.Tracking) []slack.Block {
	blocks := []slack.Block{
		slack.NewHeaderBlock(
			slack.NewTextBlockObject("plain_text", "📋 Your Task Status", false, false),
		),
		slack.NewDividerBlock(),
	}

	for _, record := range records {
		status := "⏳ In Progress"
		if record.TaskComplete {
			status = "✅ Complete"
		}

		taskBlock := slack.NewSectionBlock(
			slack.NewTextBlockObject(
				"mrkdwn",
				fmt.Sprintf("*%s*\n%s", record.TaskName, status),
				false,
				false,
			),
			nil,
			nil,
		)

		// Add context with dates
		elements := []slack.MixedElement{}
		if record.FirstMessageSent {
			elements = append(elements,
				slack.NewTextBlockObject(
					"plain_text",
					fmt.Sprintf("First notified: %s", record.FirstMessageSentAt.Format("Jan 2, 2006")),
					false,
					false,
				),
			)
		}

		if len(elements) > 0 {
			contextBlock := slack.NewContextBlock("", elements...)
			blocks = append(blocks, taskBlock, contextBlock)
		} else {
			blocks = append(blocks, taskBlock)
		}
	}

	return blocks
}

// Example of how to register these commands with proper error handling.
func (c *CommandsRefactored) RegisterCommands(bot *slacker.Slacker) {
	// User lookup command
	bot.AddCommand(&slacker.CommandDefinition{
		Command:     "user <user>",
		Description: "Look up user information",
		Examples:    []string{"user @john.doe"},
		Handler: func(ctx *slacker.CommandContext) {
			defer func() {
				if r := recover(); r != nil {
					c.log.Error().
						Err(errors.New(errors.ErrCodeInternal, "command panic")).
						Interface("panic", r).
						Msg("user command panicked")

					_, err := ctx.Response().Reply("An unexpected error occurred. Please try again.")
					if err != nil {
						c.log.Error().Err(errors.SlackError("send panic message", err)).Msg("failed to send error message")
					}
				}
			}()

			c.HandleUserLookup(ctx)
		},
	})

	// Task status command
	bot.AddCommand(&slacker.CommandDefinition{
		Command:     "status",
		Description: "Check your task status",
		Handler: func(ctx *slacker.CommandContext) {
			defer func() {
				if r := recover(); r != nil {
					c.log.Error().
						Err(errors.New(errors.ErrCodeInternal, "command panic")).
						Interface("panic", r).
						Msg("status command panicked")

					_, err := ctx.Response().Reply("An unexpected error occurred. Please try again.")
					if err != nil {
						c.log.Error().Err(errors.SlackError("send panic message", err)).Msg("failed to send error message")
					}
				}
			}()

			c.HandleTaskStatus(ctx)
		},
	})

	// Complete task command
	bot.AddCommand(&slacker.CommandDefinition{
		Command:     "complete <task>",
		Description: "Mark a task as complete",
		Examples:    []string{"complete os-update"},
		Handler: func(ctx *slacker.CommandContext) {
			defer func() {
				if r := recover(); r != nil {
					c.log.Error().
						Err(errors.New(errors.ErrCodeInternal, "command panic")).
						Interface("panic", r).
						Msg("complete command panicked")

					_, err := ctx.Response().Reply("An unexpected error occurred. Please try again.")
					if err != nil {
						c.log.Error().Err(errors.SlackError("send panic message", err)).Msg("failed to send error message")
					}
				}
			}()

			c.HandleCompleteTask(ctx)
		},
	})
}
