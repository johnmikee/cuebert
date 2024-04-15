package bot

import (
	"github.com/johnmikee/cuebert/bot/lifecycle"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/handlers"

	"github.com/johnmikee/cuebert/pkg/logger"
	slacker "github.com/shomali11/slacker/v2"
)

// Config holds the configuration for the bot.
type Config struct {
	Flags         *config.Flags
	Config        *config.Config
	Log           logger.Logger
	Interactive   Interactive
	LifeCycle     lifecycle.LifeCycle
	StatusChan    chan handlers.StatusMessage
	StatusHandler *handlers.StatusHandler
}

// Bot holds the configuration for the bot as well as interacting with the DB and Slack.
type Bot struct {
	bot           *slacker.Slacker
	flags         *config.Flags
	config        *config.Config
	interactive   Interactive
	lifecycle     lifecycle.LifeCycle
	log           logger.Logger
	statusChan    chan handlers.StatusMessage
	statusHandler *handlers.StatusHandler
}

// Interactive is an interface for the bot to interact with the DB and Slack.
type Interactive interface {
	InteractionHandler(ctx *slacker.InteractionContext)
	InteractiveHelper(ctx *slacker.InteractionContext)
	LoadInput(ctx *slacker.InteractionContext, loadType string)
	LoadProgram(triggerID string, loadType string)
	ModalSubmit(ctx *slacker.InteractionContext)
	RegisterAdminCommands()
	RegisterCommands()
	RegisterHandlers()
	RegisterHelp()
	RegisterMiddlewares()
}

// New creates a new bot.
func New(config *Config) *Bot {
	return &Bot{
		bot: slacker.NewClient(
			config.Config.SlackBotToken,
			config.Config.SlackAppToken,
			slacker.WithDebug(config.Flags.Debug),
		),
		flags:         config.Flags,
		config:        config.Config,
		interactive:   config.Interactive,
		lifecycle:     config.LifeCycle,
		log:           logger.ChildLogger("bot", &config.Log),
		statusChan:    config.StatusChan,
		statusHandler: config.StatusHandler,
	}
}

// Bot returns the bot.
func (b *Bot) Bot() *slacker.Slacker {
	return b.bot
}
