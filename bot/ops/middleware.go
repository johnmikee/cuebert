package ops

import (
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker/v2"
)

// Middlewares is a type for the different middlewares that can be registered.
type Middlewares string

const (
	Authorization      Middlewares = "AuthorizationMiddleware"
	LoggingInteraction Middlewares = "LoggingInteractionMiddleware"
)

// Middleware is a struct that holds the bot, users, and logger.
type Middleware struct {
	Bot   *slacker.Slacker
	Users []string
	Log   logger.Logger
}

// AuthorizationMiddleware checks if a user is authorized to use a given command.
func AuthorizationMiddleware(authorizedUserNames []string) slacker.CommandMiddlewareHandler {
	return func(next slacker.CommandHandler) slacker.CommandHandler {
		return func(ctx *slacker.CommandContext) {
			if helpers.Contains(authorizedUserNames, ctx.Event().UserID) {
				next(ctx)
			}
		}
	}
}

// LoggingInteractionMiddleware logs all interactions.
func LoggingInteractionMiddleware(log logger.Logger) slacker.InteractionMiddlewareHandler {
	return func(next slacker.InteractionHandler) slacker.InteractionHandler {
		return func(ctx *slacker.InteractionContext) {
			log.Trace().
				Str("user", ctx.Callback().User.ID).
				Str("callback", ctx.Callback().CallbackID).
				Str("channel", ctx.Callback().Channel.ID).
				Interface("action", ctx.Callback()).
				Msg("interaction received")
			next(ctx)
		}
	}
}

// RegisterMiddleWares registers the middlewares.
func RegisterMiddleWares(m *Middleware, opts ...Middlewares) {
	for _, o := range opts {
		switch o {
		case Authorization:
			m.Bot.AddCommandMiddleware(AuthorizationMiddleware(m.Users))
		case LoggingInteraction:
			m.Bot.AddInteractionMiddleware(LoggingInteractionMiddleware(m.Log))
		}
	}
}
