package main

import (
	"context"
	"fmt"
	"time"

	"strings"

	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

// Bot holds the configuration for the bot as well
// as interacting with the DB and Slack.
type Bot struct {
	cfg    *CuebertConfig
	bot    *slacker.Slacker
	idp    idp.Provider
	mdm    mdm.Provider
	tables *Tables

	log logger.Logger
	db  *db.Config

	commands []Commands

	reloadSignal  chan struct{}
	statusChan    chan StatusMessage
	statusHandler *StatusHandler
	startSignal   chan struct{}
	stopSignal    chan struct{}
	isRunning     bool
}

// Commands holds the usage and definition for a command
// that the bot will respond to.
type Commands struct {
	usage string
	def   *slacker.CommandDefinition
}

func (b *Bot) respondUpdater(msg string, err error) {
	status := b.statusHandler.GetStatus()

	status.Respond = &BotStatus{
		Name:    "respond",
		Error:   err,
		Message: msg,
		Time:    time.Now().Format(time.RFC3339),
	}

	b.statusHandler.SetStatus(status)
}

func (b *Bot) respond() {
	go b.respondUpdater("starting up", nil)

	b.log.Info().Msg("starting responder...")
	// register user info
	b.getSelfInfo()
	// register device info
	b.deviceInfo()
	// register exception info
	b.exceptionHelp()
	// reqister reminder info
	b.reminderHelp()

	for _, c := range b.commands {
		b.bot.Command(c.usage, c.def)
	}

	// add the admin commands
	b.bot.Command("add exclusion", b.addExclusion())
	b.bot.Command("get report <opt>", b.requestReport())
	b.bot.Command("stop cuebert", b.stopper())
	b.bot.Command("start cuebert", b.initSettings())
	b.bot.Command("update cuebert", b.updateConfig())
	b.bot.Command("get users info <opt> {input}", b.getUsersInfo())
	// handle messages needing an acknowledgement
	b.bot.Interactive(func(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
		if callback.Type == slack.InteractionTypeInteractionMessage {
			if callback.OriginalMessage.User != b.cfg.SlackBotID {
				b.log.Trace().
					Str("callback_bot_id", callback.OriginalMessage.BotID).
					Str("our_bot_id", b.cfg.SlackBotID).
					Interface("callback", callback).
					Msg("ignoring callback from another bot")
				return
			} else if callback.Type == slack.InteractionTypeViewSubmission {
				if callback.View.BotID != b.cfg.SlackBotID {
					b.log.Trace().
						Str("callback_bot_id", callback.View.BotID).
						Str("our_bot_id", b.cfg.SlackBotID).
						Interface("callback", callback).
						Msg("ignoring callback from another bot")
					return
				}
			}
		}
		switch callback.CallbackID {
		case "ack_it":
			b.ack(s, event, callback)
		case "remind_me_question":
			b.reminderRequested(s, event, callback)
		case "exception_question", "exclusion_add_question":
			b.exceptionRequested(s, event, callback)
		case "exclusion_approver":
			b.exceptionRequestDecision(s, event, callback)
		case "stop_cuebert_request":
			b.stopApprover(s, event, callback)
		case "stop_cuebert_approve":
			b.stopSubmit(s, event, callback)
		case "stop_cuebert_override":
			b.overrideSubmit(s, event, callback)
		case START, RELOAD:
			b.updateRequested(s, event, callback)
		case "":
			switch callback.Type {
			case slack.InteractionTypeBlockActions:
				b.helpAck(s, event, callback)
			case slack.InteractionTypeViewSubmission:
				// these modals dont have a callback id for reasons that are
				// escaping me. pass it to the modalSubmit for further validation
				b.modalSubmit(s, event, callback)
			default:
				b.log.Trace().Str("callback", callback.CallbackID).Msg("not an interaction we are handling")
			}
		default:
			b.log.Trace().Str("callback", callback.CallbackID).Msg("not an interaction we are handling")

		}
	})

	// set the custom help response
	b.bot.Help(b.help(b.cfg.authUsers))

	// set the default handler for any messages to cuebert that do not match our inputs.
	b.bot.DefaultCommand(func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
		if botCtx.Event().User != b.cfg.SlackBotID {
			// Data is the raw event data returned from slack. By switching over Type, we can cast
			// this into a slackevents *Event struct.
			//
			// Right now we are only using the message_deleted event to ignore but this may
			// expand in the future so leaving the switch here.
			switch botCtx.Event().Type {
			case "message":
				event := botCtx.Event().Data.(*slackevents.MessageEvent)
				if event.SubType == "message_deleted" {
					b.log.Trace().Msg("ignoring message_deleted event")
					return
				}
			default:
				// no-op
				b.log.Trace().Str("event_type", botCtx.Event().Type).Msg("ignoring event type")
			}
			err := defaultReply(response)
			if err != nil {
				// this will get polluted with a specific error
				// unless we ignore it.
				if err.Error() != "not_in_channel" {
					b.log.Trace().AnErr("default_command", err).Send()
				}
			}
		} else {
			b.log.Info().Str("user", botCtx.Event().User).Msg("ignoring message from bot")
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

func defaultReply(response slacker.ResponseWriter) error {
	return response.Reply("Sorry, I didn't understand that request. Try `help`.")
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

	channelID, timestamp, err := b.bot.Client().PostMessage(channel, message)
	if err != nil {
		b.log.Err(err).Msg("posting message")
	}

	b.log.Trace().
		Str("message", msg).
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Msg("message sent")
}
