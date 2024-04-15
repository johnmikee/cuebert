package bot

import (
	"context"
	"time"

	"github.com/johnmikee/cuebert/handlers"
)

// Respond starts the bot.
func (b *Bot) Respond() {
	go b.respondUpdater("starting up", nil)

	b.log.Info().Msg("starting responder...")
	// register the middleware(s)
	b.interactive.RegisterMiddlewares()
	// register the handler(s)
	b.interactive.RegisterHandlers()
	// register the custom help command
	b.interactive.RegisterHelp()
	// register default commands
	b.interactive.RegisterCommands()
	// register the admin commands
	b.interactive.RegisterAdminCommands()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := b.bot.Listen(ctx)
	if err != nil {
		b.log.Info().AnErr("error", err).Msg("error listening")
		b.respondUpdater("failed to respond", err)
	}
}

func (b *Bot) respondUpdater(msg string, err error) {
	status := b.statusHandler.GetStatus()

	status.Respond = &handlers.BotStatus{
		Name:    "respond",
		Error:   err,
		Message: msg,
		Time:    time.Now().Format(time.RFC3339),
	}

	b.statusHandler.SetStatus(status)
}
