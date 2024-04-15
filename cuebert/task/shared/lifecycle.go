package shared

import (
	"fmt"
	"strings"

	"github.com/johnmikee/cuebert/bot"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// loadPrompt leads us to the modal to update cuebert
func (c *Shared) LoadPrompt(user, title, callback string) {
	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       title,
			CallbackID: c.constFormat(callback),
			YesName:    c.constFormat(UpdateConfigYes),
			YesText:    "Yes",
			YesValue:   c.constFormat(UpdateConfigYes),
			YesStyle:   "primary",
			NoName:     UpdateConfigNo,
			NoText:     "No",
			NoValue:    UpdateConfigNo,
			NoStyle:    "danger",
			Channel:    user,
			Msg:        "updating config prompt",
		},
	)
}

func (c *Shared) Launch(ctx *slacker.InteractionContext) {
	c.log.Info().Interface("callback", ctx.Callback()).Msg("launching")
	if ctx.Callback().Type == slack.InteractionTypeInteractionMessage {
		action := ctx.Callback().ActionCallback.AttachmentActions[0]
		switch action.Name {
		case c.constFormat(UpdateConfigYes):
			c.LoadProgram(
				&Loader{
					TriggerID: ctx.Callback().TriggerID,
					LoadType:  c.constFormat(Start),
					Tables:    Tables(c.flags.TaskImplementation, TaskType(strings.ToLower(c.flags.Task))),
				},
			)
		case UpdateConfigNo:
			c.lifecycle.Start()
		}
	} else {
		c.lifecycle.Start()
	}

	_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
}

// OverrideSubmit acknowledges the stop request and stops cuebert
func (c *Shared) OverrideSubmit(ctx *slacker.InteractionContext) {
	_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

	_, _, err := c.bot.Bot().SlackClient().PostMessage(
		c.config.SlackAlertChannel,
		slack.MsgOptionText(
			fmt.Sprintf(
				"Judge, Jury, and Executioner: Cuebert stopped by <@%s>. :white_check_mark:", ctx.Callback().User.ID),
			false),
	)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("posting message")
	}

	c.lifecycle.Stop()
	c.log.Info().Msg("cuebert stop approved")
}

// note who submitted the stop request and present an approval message
func (c *Shared) StopApprover(ctx *slacker.InteractionContext) {
	_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

	user := ctx.Callback().User.ID

	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       fmt.Sprintf("Do you approve stopping cuebert? Requested by: <@%s>", user),
			Fallback:   user,
			CallbackID: c.constFormat(StopCuebertApproval),
			YesName:    ApproveStop,
			YesText:    "Approve",
			YesValue:   ApproveStop,
			YesStyle:   "danger",
			NoName:     DenyStop,
			NoText:     "Deny",
			NoValue:    DenyStop,
			NoStyle:    "primary",
			Channel:    c.config.SlackAlertChannel,
			Msg:        "request for approving cuebert stop",
		},
	)
}

// sometimes you just gotta break the rules. this allows us to bypass the approval process.
// maybe cuebert went sentient. maybe we just need to stop cuebert now.
func (c *Shared) StopApprovalOverride() {
	c.bot.ModalGateway(
		&bot.ModalGateway{
			Text:       "Are you sure you want to bypass the approval process and stop cuebert?",
			CallbackID: c.constFormat(StopCuebertOverride),
			YesName:    BypassOverrideYes,
			YesText:    "Yes",
			YesValue:   BypassOverrideYes,
			YesStyle:   "danger",
			NoName:     BypassOverrideNo,
			NoText:     "Cancel",
			NoValue:    BypassOverrideNo,
			NoStyle:    "primary",
			Channel:    c.config.SlackAlertChannel,
			Msg:        "request for bypassing approval process and stopping cuebert",
		},
	)
}

func (c *Shared) StopSubmit(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0]
	approver := ctx.Callback().User.ID
	requester := ctx.Callback().OriginalMessage.Attachments[0].Fallback

	c.log.Trace().
		Str("approver", approver).
		Str("requester", requester).
		Msg("stop submit")

	switch action.Value {
	case ApproveStop:
		_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
		if approver != requester {
			c.log.Info().Msg("cuebert stop approved")
			c.lifecycle.Stop()
		} else {
			c.StopApprovalOverride()
		}

	case DenyStop:
		_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

		_, _, err := c.bot.Bot().SlackClient().PostMessage(ctx.Callback().User.ID,
			slack.MsgOptionText("Cuebert stop denied :white_check_mark:", false))
		if err != nil {
			c.log.Err(err).Msg("posting message")
		}

		c.log.Info().Msg("cuebert stop denied")
	}
}
