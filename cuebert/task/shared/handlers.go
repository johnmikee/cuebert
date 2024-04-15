package shared

import (
	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/shomali11/slacker/v2"
)

func (c *Shared) RegisterHandlers() {
	ops.RegisterHandlers(
		&ops.Handle{
			Bot:          c.bot.Bot(),
			Log:          c.log,
			Handlers:     []ops.Handlers{ops.UnsupportedCommand, ops.OnDisconnected},
			AlertChannel: c.config.SlackAlertChannel,
		},
	)

	c.bot.Bot().UnsupportedInteractionHandler(func(ctx *slacker.InteractionContext) {
		c.InteractionHandler(ctx)
	})
}

func (c *Shared) RegisterMiddlewares() {
}

func (c *Shared) AuthorizationMiddleware() slacker.CommandMiddlewareHandler {
	return ops.AuthorizationMiddleware(c.flags.AuthUsers)
}
