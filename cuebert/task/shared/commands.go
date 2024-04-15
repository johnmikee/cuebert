package shared

import (
	"encoding/json"

	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/shomali11/slacker/v2"
)

// CheckComplianceStatus is used for both tasks to make a call to see if the user
// has either updated their device or completed the lesson.
//
// This is a way to self-service if the user has gone beyond the deadline and wants
// to force a check to unblock access before the next scheduled check.
func (c *Shared) CheckComplianceStatus() {
	definition := &slacker.CommandDefinition{
		Command:     "check my compliance",
		Description: "Check compliance status",
		Examples:    []string{"check my compliance"},
		Middlewares: nil,
		Handler: func(ctx *slacker.CommandContext) {
			c.wantCheck("Would you like to check your compliance status?", ctx)
		},
	}
	c.bot.Bot().AddCommand(definition)
}

// Exclusion commands

// AddExclusion adds a device to be excluded. Only authorized users can add exclusions.
func (c *Shared) AddExclusion() {
	definition := &slacker.CommandDefinition{
		Command:     "add exclusion",
		Description: "Add a device to be excluded",
		Examples:    []string{"add exclusion"},
		Middlewares: []slacker.CommandMiddlewareHandler{ops.AuthorizationMiddleware(c.flags.AuthUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			c.wantExclusion(
				&exclusionary{
					ctx:        ctx,
					text:       "Would you like to add an exclusion?",
					callbackID: ExclusionAddQuestion,
					yes:        ExclusionAddYes,
					no:         ExclusionAddNo,
					msg:        "exclusion add request",
				},
			)
		},
	}

	c.bot.Bot().AddCommand(definition)
}

// ExclusionHelp invokes the exclusion modal users can use to request an exclusion
func (c *Shared) ExclusionHelp() {
	definition := &slacker.CommandDefinition{
		Command:     "request exclusion",
		Description: "Request an exclusion",
		Examples:    []string{"request exclusion"},
		Middlewares: nil,
		Handler: func(ctx *slacker.CommandContext) {
			c.wantExclusion(
				&exclusionary{
					ctx:        ctx,
					text:       "Would you like to request an exclusion?",
					callbackID: ExclusionQuestion,
					yes:        YesExclusion,
					no:         NoExclusion,
					msg:        "exclusion requested",
				},
			)
		},
	}
	c.bot.Bot().AddCommand(definition)
}

// Default Bot Commands

// InitSettings allows authorized users to set the initial configuration for the bot.
func (c *Shared) InitSettings() {
	definition := &slacker.CommandDefinition{
		Command:     "start cuebert",
		Description: "Set the initial configuration for the bot",
		Examples:    []string{"start cuebert"},
		Middlewares: []slacker.CommandMiddlewareHandler{ops.AuthorizationMiddleware(c.flags.AuthUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			c.LoadPrompt(
				ctx.Event().UserID,
				"Do you want to update the configuration before starting?",
				UpdateBeforeLaunch,
			)
		},
	}
	c.bot.Bot().AddCommand(definition)
}

// MessageTest allows us to stop cuebert from running via slack. requires secondary approval
func (c *Shared) MessageTest() {
	definition := &slacker.CommandDefinition{
		Command:     "test messaging",
		Description: "Test messaging",
		Middlewares: []slacker.CommandMiddlewareHandler{ops.AuthorizationMiddleware(c.flags.AuthUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			c.StopRequest(ctx.Event().UserID)
		},
		HideHelp: false,
	}
	c.bot.Bot().AddCommand(definition)
}

// Stopper allows us to stop cuebert from running via slack. requires secondary approval
func (c *Shared) Stopper() {
	definition := &slacker.CommandDefinition{
		Command:     "stop cuebert",
		Description: "Stop cuebert",
		Middlewares: []slacker.CommandMiddlewareHandler{ops.AuthorizationMiddleware(c.flags.AuthUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			c.StopRequest(ctx.Event().UserID)
		},
		HideHelp: false,
	}
	c.bot.Bot().AddCommand(definition)
}

// updateConfig allows authorized users to update the configuration for the bot.
func (c *Shared) UpdateConfig() {
	definition := &slacker.CommandDefinition{
		Command:     "update cuebert",
		Description: "Update the configuration",
		Examples:    []string{"update cuebert"},
		Middlewares: []slacker.CommandMiddlewareHandler{ops.AuthorizationMiddleware(c.flags.AuthUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			c.LoadPrompt(ctx.Event().UserID, "Do you want to update the configuration?", Reload)
		},
	}
	c.bot.Bot().AddCommand(definition)
}

// getStatus returns the status of the bot
func (c *Shared) GetStatus(status *handlers.StatusHandler) {
	definition := &slacker.CommandDefinition{
		Command:     "get status",
		Description: "Get the status of the bot",
		Examples:    []string{"get status"},
		Middlewares: []slacker.CommandMiddlewareHandler{ops.AuthorizationMiddleware(c.flags.AuthUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			st := status.GetStatus()
			status, err := json.Marshal(st)
			if err != nil {
				c.log.Debug().AnErr("marshalling status", err).Send()
				return
			}
			_, err = ctx.Response().Reply(string(status))
			if err != nil {
				c.log.Debug().AnErr("sending status", err).Send()
			}
		},
	}
	c.bot.Bot().AddCommand(definition)
}

// reminderHelp adds the reminder command to the bot
func (c *Shared) ReminderHelp() {
	var title string
	switch c.flags.TaskImplementation {
	case string(implementation.OS):
		title = "Would you like a reminder to update your device?"
	default:
		title = "Would you like a reminder to finish the lesson?"
	}
	definition := &slacker.CommandDefinition{
		Command:     "request reminder",
		Description: "Request a reminder to update",
		Examples:    []string{"request reminder"},
		Middlewares: nil,
		Handler: func(ctx *slacker.CommandContext) {
			c.wantAReminder(title, ctx)
		},
	}
	c.bot.Bot().AddCommand(definition)
}

func (c *Shared) MessageTestRequest(channel string) {
	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       "Proceed with test message",
			Fallback:   channel,
			CallbackID: c.constFormat(MessageTestRequest),
			YesName:    MessageRequestYes,
			YesText:    "Yes",
			YesValue:   MessageRequestYes,
			YesStyle:   "danger",
			NoName:     MessageRequestNo,
			NoText:     "No",
			NoValue:    MessageRequestNo,
			NoStyle:    "primary",
			Channel:    channel,
			Msg:        "request for approving test message",
		},
	)
}

func (c *Shared) StopRequest(channel string) {
	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       "Are you sure you want to stop cuebert?",
			Fallback:   channel,
			CallbackID: c.constFormat(StopCuebertRequest),
			YesName:    StopCuebertYes,
			YesText:    "Yes",
			YesValue:   StopCuebertYes,
			YesStyle:   "danger",
			NoName:     StopCuebertNo,
			NoText:     "No",
			NoValue:    StopCuebertNo,
			NoStyle:    "primary",
			Channel:    channel,
			Msg:        "request for approving cuebert stop",
		},
	)
}

type exclusionary struct {
	ctx        *slacker.CommandContext
	text       string
	callbackID string
	yes        string
	no         string
	msg        string
}

func (c *Shared) wantExclusion(e *exclusionary) {
	c.log.Debug().Str("callbackID", e.callbackID).Send()
	c.log.Info().Interface("event", e.ctx.Event()).Send()
	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       e.text,
			CallbackID: c.constFormat(e.callbackID),
			YesName:    e.yes,
			YesText:    "Yes",
			YesValue:   e.yes,
			YesStyle:   "primary",
			NoName:     e.no,
			NoText:     "No",
			NoValue:    e.no,
			NoStyle:    "danger",
			Channel:    e.ctx.Event().UserID,
			Msg:        e.msg,
		},
	)
}

func (c *Shared) wantCheck(title string, ctx *slacker.CommandContext) {
	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       title,
			CallbackID: c.constFormat(CheckComplianceStatus),
			YesName:    YesCheck,
			YesText:    "Yes",
			YesValue:   YesCheck,
			YesStyle:   "primary",
			NoName:     NoCheck,
			NoText:     "No",
			NoValue:    NoCheck,
			NoStyle:    "danger",
			Channel:    ctx.Event().UserID,
			Msg:        "compliance check requested",
		},
	)
}

func (c *Shared) wantAReminder(title string, ctx *slacker.CommandContext) {
	c.log.Info().Interface("event", ctx.Event()).Send()
	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       title,
			CallbackID: c.constFormat(RemindMeQuestion),
			YesName:    RemindMe,
			YesText:    "Yes",
			YesValue:   RemindMe,
			YesStyle:   "primary",
			NoName:     DontRemindMe,
			NoText:     "No",
			NoValue:    DontRemindMe,
			NoStyle:    "danger",
			Channel:    ctx.Event().UserID,
			Msg:        "reminder requested",
		},
	)
}
