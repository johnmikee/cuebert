package bot

import (
	"context"
	"fmt"
	"time"

	"strings"

	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/cuebert/tables"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/idp"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// Bot holds the configuration for the bot as well
// as interacting with the DB and Slack.
type Bot struct {
	bot           *slacker.Slacker
	tables        *tables.Config
	cfg           *Cfg
	log           logger.Logger
	lifecycle     LifeCycle
	method        Method
	statusChan    chan handlers.StatusMessage
	statusHandler *handlers.StatusHandler
}

type Method interface {
	Messaging
	Updates
}

type LifeCycle interface {
	Start()
	Stop()
	Update()
}

type Messaging interface {
	FirstMessage() string
	ReminderMessage(rp *ReminderPayload) error
}

type Updates interface {
	UserUpdate(ctx *slacker.InteractionContext)
	UserUpdateModalSubmit(base *bot.Info, values map[string]map[string]slack.BlockAction)
}

// Config holds the configuration for the bot.
type Config struct {
	Cfg           *Cfg
	SlackBotToken string
	SlackAppToken string
	DB            *db.DB
	IDP           idp.Provider
	MDM           mdm.Provider
	Log           logger.Logger
	LifeCycle     LifeCycle
	Method        Method
	Tables        *tables.Config
	StatusChan    chan handlers.StatusMessage
	StatusHandler *handlers.StatusHandler
}

// New creates a new bot.
func New(config *Config) *Bot {
	return &Bot{
		bot:           slacker.NewClient(config.SlackBotToken, config.SlackAppToken, slacker.WithDebug(false)),
		cfg:           config.Cfg,
		log:           logger.ChildLogger("bot", &config.Log),
		lifecycle:     config.LifeCycle,
		method:        config.Method,
		tables:        config.Tables,
		statusHandler: config.StatusHandler,
		statusChan:    config.StatusChan,
	}
}

func (b *Bot) Client() *slack.Client {
	return b.bot.SlackClient()
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

func (b *Bot) Respond() {
	go b.respondUpdater("starting up", nil)

	b.log.Info().Msg("starting responder...")

	// register the custom help command
	b.bot.Help(b.help())

	// register user commands
	b.getSelfInfo()
	b.deviceInfo()
	b.exclusionHelp()
	b.reminderHelp()

	// register the admin commands
	b.initSettings()
	b.stopper()
	b.updateConfig()
	b.addExclusion()
	b.requestReport()
	b.getUsersInfo()
	b.updateUserInfoInteractive()

	// register the interactive commands and middleware
	b.bot.AddInteractionMiddleware(b.loggingInteractionMiddleware())
	b.bot.AddInteraction(
		&slacker.InteractionDefinition{
			BlockID: RemindMeQuestion,
			Handler: b.reminderRequested,
		},
	)
	b.bot.GetJobs()

	b.bot.AddInteraction(
		&slacker.InteractionDefinition{
			BlockID: ExclusionAddQuestion,
			Handler: b.exclusionRequested,
		},
	)

	b.bot.AddInteraction(
		&slacker.InteractionDefinition{
			BlockID: ExclusionQuestion,
			Handler: b.exclusionRequested,
		},
	)
	// using this to sort modals for the time being.
	b.bot.UnsupportedInteractionHandler(func(ctx *slacker.InteractionContext) {
		b.interactive(ctx)
	})

	// TODO: webhook this error
	b.bot.OnDisconnected(func(event socketmode.Event) {
		b.log.Info().Interface("event", event).Msg("disconnected")
	})

	b.bot.UnsupportedCommandHandler(func(ctx *slacker.CommandContext) {
		if ctx.Event().UserProfile != nil {
			_, err := ctx.Response().Reply("Sorry, I didn't understand that request. Try `help`.")
			if err != nil {
				b.log.Err(err).Send()
			}
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := b.bot.Listen(ctx)
	if err != nil {
		b.log.Err(err).Send()
		b.respondUpdater("failed to respond", err)
	}
}

func (b *Bot) interactive(ctx *slacker.InteractionContext) {
	switch ctx.Callback().CallbackID {
	case AckIT:
		b.ack(ctx)
	case ExclusionApprover:
		b.exclusionRequestDecision(ctx)
	case StopCuebertRequest:
		b.stopApprover(ctx)
	case StopCuebertApproval:
		b.stopSubmit(ctx)
	case StopCuebertOverride:
		b.overrideSubmit(ctx)
	case Start, Reload, UpdateUserQuestion:
		b.interactiveHelper(ctx)
	case "":
		switch ctx.Callback().Type {
		// case slack.InteractionTypeBlockActions:
		// 	b.helpAck(ctx)
		case slack.InteractionTypeViewSubmission:
			// these modals dont have a callback id for reasons that are
			// escaping me. pass it to the modalSubmit for further validation
			b.modalSubmit(ctx)
		default:
			b.log.Trace().Str("callback", ctx.Callback().CallbackID).Msg("not an interaction we are handling")
		}
	default:
		b.log.Trace().Str("callback", ctx.Callback().CallbackID).Msg("not an interaction we are handling")

	}
}

func (b *Bot) loggingInteractionMiddleware() slacker.InteractionMiddlewareHandler {
	return func(next slacker.InteractionHandler) slacker.InteractionHandler {
		return func(ctx *slacker.InteractionContext) {
			b.log.Trace().
				Str("user", ctx.Callback().User.ID).
				Str("callback", ctx.Callback().CallbackID).
				Str("channel", ctx.Callback().Channel.ID).
				Interface("action", ctx.Callback()).
				Msg("interaction received")
			next(ctx)
		}
	}
}

func fuzzyMatchNonOpt(opt string, opts []string) string {
	maybeMatch := fuzzy.RankFind(opt, opts)
	var msg string

	if len(maybeMatch) > 0 {
		msg = fmt.Sprintf("Sorry `%s` isnt a valid option. Did you mean `%s`?\n\nValid options are:\n%s",
			opt,
			maybeMatch[0].Target,
			strings.Join(opts, " \n"))

	} else {
		msg = fmt.Sprintf("Sorry `%s` isnt a valid option.\n\nValid options are:\n%s",
			opt,
			strings.Join(opts, " \n"))
	}

	return msg
}

func (b *Bot) sendAttachment(attachment slack.Attachment, channel, msg string) {
	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := b.bot.SlackClient().PostMessage(channel, message)
	if err != nil {
		b.log.Err(err).Msg("posting message")
	}

	b.log.Trace().
		Str("message", msg).
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Msg("message sent")
}
