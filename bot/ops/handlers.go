package ops

import (
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Handlers is a type for the different handlers that can be registered
type Handlers string

const (
	UnsupportedCommand Handlers = "UnsupportedCommand"
	OnDisconnected     Handlers = "OnDisconnect"
)

// Handle is a struct that holds the bot, handlers, logger, and alert channel
type Handle struct {
	Bot          *slacker.Slacker
	Handlers     []Handlers
	Log          logger.Logger
	AlertChannel string
}

// DefaultHandlers registers the default handlers
func DefaultHandlers(c *Handle) {
	c.UnsupportedCommandHandler()
}

// RegisterHandlers registers the handlers
func RegisterHandlers(c *Handle) {
	for _, h := range c.Handlers {
		switch h {
		case UnsupportedCommand:
			c.UnsupportedCommandHandler()
		case OnDisconnected:
			c.DisconnectedHandler()
		}
	}
}

// DisconnectedHandler registers the disconnected handler that runs when the bot disconnects from Slack
func (h *Handle) DisconnectedHandler() {
	h.Bot.OnDisconnected(func(event socketmode.Event) {
		h.Log.Info().Interface("event", event).Msg("disconnected")
		_, _, err := h.Bot.SlackClient().PostMessage(
			h.AlertChannel,
			slack.MsgOptionText("Cuebert has disconnected from Slack", false))
		if err != nil {
			h.Log.Info().AnErr("error", err).Msg("registering disconnected handler")
		}
	})
}

// UnsupportedCommandHandler registers the unsupported command handler that runs when the bot receives an unsupported command
func (h *Handle) UnsupportedCommandHandler() {
	h.Bot.UnsupportedCommandHandler(func(ctx *slacker.CommandContext) {
		if ctx.Event().UserProfile != nil {
			_, err := ctx.Response().Reply("Sorry, I didn't understand that request. Try `help`.")
			if err != nil {
				h.Log.Info().AnErr("error", err).Msg("registering unsupported command handler")
			}
		}
	},
	)
}
