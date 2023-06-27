package main

import (
	"fmt"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// initSettings allows authorized users to set the initial configuration for the bot.
func (b *Bot) initSettings() *slacker.CommandDefinition {
	return &slacker.CommandDefinition{
		Description: "Set the initial configuration for the bot",
		Examples:    []string{"init settings"},
		AuthorizationFunc: func(botCtx slacker.BotContext, request slacker.Request) bool {
			return helpers.Contains(b.cfg.authUsers, botCtx.Event().User)
		},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			b.loadPrompt(botCtx.Event().User, START)
		},
		HideHelp: false,
	}
}

// overrideSubmit acknowledges the stop request and stops cuebert
func (b *Bot) overrideSubmit(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	s.SocketMode().Ack(*event.Request)

	_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

	_, _, err := s.Client().PostMessage(b.cfg.SlackAlertChannel,
		slack.MsgOptionText(
			fmt.Sprintf(
				"Judge, Jury, and Executioner: Cuebert stopped by <@%s>. :white_check_mark:", callback.User.ID),
			false))
	if err != nil {
		b.log.Err(err).Msg("posting message")
	}

	b.stop()
	b.log.Info().Msg("cuebert stop approved")
}

// stopper allows us to stop cuebert from running via slack. requires secondary approval
func (b *Bot) stopper() *slacker.CommandDefinition {
	return &slacker.CommandDefinition{
		Description: "Stop cuebert",
		AuthorizationFunc: func(botCtx slacker.BotContext, request slacker.Request) bool {
			return helpers.Contains(b.cfg.authUsers, botCtx.Event().User)
		},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			b.stopRequest(botCtx.Event().User)
		},
		HideHelp: false,
	}
}

// updateConfig allows authorized users to update the configuration for the bot.
func (b *Bot) updateConfig() *slacker.CommandDefinition {
	return &slacker.CommandDefinition{
		Description: "Update the configuration",
		Examples:    []string{"update config"},
		AuthorizationFunc: func(botCtx slacker.BotContext, request slacker.Request) bool {
			return helpers.Contains(b.cfg.authUsers, botCtx.Event().User)
		},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			b.loadPrompt(botCtx.Event().User, RELOAD)
		},
		HideHelp: false,
	}
}

// updateRequested takes us to the modal to update cuebert
func (b *Bot) updateRequested(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	if callback.Type != slack.InteractionTypeInteractionMessage {
		return
	}

	action := callback.ActionCallback.AttachmentActions[0]

	if action.Name == "yes_update_config" {
		b.loadProgram(callback.TriggerID, callback.CallbackID)
	}
	_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)
	s.SocketMode().Ack(*event.Request)
}

// note who submitted the stop request and present an approval message
func (b *Bot) stopApprover(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)
	s.SocketMode().Ack(*event.Request)

	user := callback.User.ID

	b.modalGateway(
		&modalGateway{
			text:       fmt.Sprintf("Do you approve stopping cuebert? Requested by: <@%s>", user),
			fallback:   user,
			callbackID: "stop_cuebert_approve",
			yesName:    "approve_stop",
			yesText:    "Approve",
			yesValue:   "approve_stop",
			yesStyle:   "danger",
			noName:     "deny_stop",
			noText:     "Deny",
			noValue:    "deny_stop",
			noStyle:    "primary",
			channel:    b.cfg.SlackAlertChannel,
			msg:        "request for approving cuebert stop",
		},
	)
}

// sometimes you just gotta break the rules. this allows us to bypass the approval process.
// maybe cuebert went sentient. maybe we just need to stop cuebert now.
func (b *Bot) stopApprovalOverride() {
	b.modalGateway(&modalGateway{
		text:       "Are you sure you want to bypass the approval process and stop cuebert?",
		callbackID: "stop_cuebert_override",
		yesName:    "bypass_override_yes",
		yesText:    "Yes",
		yesValue:   "bypass_override_yes",
		yesStyle:   "danger",
		noName:     "bypass_override_no",
		noText:     "Cancel",
		noValue:    "bypass_override_no",
		noStyle:    "primary",
		channel:    b.cfg.SlackAlertChannel,
		msg:        "request for bypassing approval process and stopping cuebert",
	})
}

// send the stop request
func (b *Bot) stopRequest(user string) {
	b.modalGateway(&modalGateway{
		text:       "Are you sure you want to stop cuebert?",
		callbackID: "stop_cuebert_request",
		yesName:    "stop_request_yes",
		yesText:    "Yes",
		yesValue:   "stop_request_yes",
		yesStyle:   "danger",
		noName:     "stop_request_no",
		noText:     "No",
		noValue:    "stop_request_no",
		noStyle:    "primary",
		channel:    user,
		msg:        "request for approving cuebert stop",
	})
}

// parse the input of stopApprover to see if we should stop cuebert
func (b *Bot) stopSubmit(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	action := callback.ActionCallback.AttachmentActions[0]
	approver := callback.User.ID
	requester := callback.OriginalMessage.Attachments[0].Fallback

	b.log.Trace().
		Str("approver", approver).
		Str("requester", requester).
		Msg("stop submit")

	switch action.Value {
	case "approve_stop":
		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)
		s.SocketMode().Ack(*event.Request)
		if approver != requester {
			b.log.Info().Msg("cuebert stop approved")
			b.stop()
		} else {
			b.stopApprovalOverride()
		}

	case "deny_stop":
		s.SocketMode().Ack(*event.Request)

		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

		_, _, err := s.Client().PostMessage(callback.User.ID,
			slack.MsgOptionText("Cuebert stop denied :white_check_mark:", false))
		if err != nil {
			b.log.Err(err).Msg("posting message")
		}

		b.log.Info().Msg("cuebert stop denied")
	}
}
