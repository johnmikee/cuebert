package bot

import (
	"fmt"

	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// overrideSubmit acknowledges the stop request and stops cuebert
func (b *Bot) overrideSubmit(ctx *slacker.InteractionContext) {
	_, _, _ = b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

	_, _, err := b.bot.SlackClient().PostMessage(
		b.cfg.slackAlertChannel,
		slack.MsgOptionText(
			fmt.Sprintf(
				"Judge, Jury, and Executioner: Cuebert stopped by <@%s>. :white_check_mark:", ctx.Callback().User.ID),
			false),
	)
	if err != nil {
		b.log.Err(err).Msg("posting message")
	}

	b.lifecycle.Stop()
	b.log.Info().Msg("cuebert stop approved")
}

// note who submitted the stop request and present an approval message
func (b *Bot) stopApprover(ctx *slacker.InteractionContext) {
	_, _, _ = b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

	user := ctx.Callback().User.ID

	b.modalGateway(
		&modalGateway{
			text:       fmt.Sprintf("Do you approve stopping cuebert? Requested by: <@%s>", user),
			fallback:   user,
			callbackID: StopCuebertApproval,
			yesName:    ApproveStop,
			yesText:    "Approve",
			yesValue:   ApproveStop,
			yesStyle:   "danger",
			noName:     DenyStop,
			noText:     "Deny",
			noValue:    DenyStop,
			noStyle:    "primary",
			channel:    b.cfg.slackAlertChannel,
			msg:        "request for approving cuebert stop",
		},
	)
}

// sometimes you just gotta break the rules. this allows us to bypass the approval process.
// maybe cuebert went sentient. maybe we just need to stop cuebert now.
func (b *Bot) stopApprovalOverride() {
	b.modalGateway(
		&modalGateway{
			text:       "Are you sure you want to bypass the approval process and stop cuebert?",
			callbackID: StopCuebertOverride,
			yesName:    BypassOverrideYes,
			yesText:    "Yes",
			yesValue:   BypassOverrideYes,
			yesStyle:   "danger",
			noName:     BypassOverrideNo,
			noText:     "Cancel",
			noValue:    BypassOverrideNo,
			noStyle:    "primary",
			channel:    b.cfg.slackAlertChannel,
			msg:        "request for bypassing approval process and stopping cuebert",
		},
	)
}

// send the stop request
func (b *Bot) stopRequest(user string) {
	b.modalGateway(
		&modalGateway{
			text:       "Are you sure you want to stop cuebert?",
			callbackID: StopCuebertRequest,
			yesName:    StopCuebertYes,
			yesText:    "Yes",
			yesValue:   StopCuebertYes,
			yesStyle:   "danger",
			noName:     StopCuebertNo,
			noText:     "No",
			noValue:    StopCuebertNo,
			noStyle:    "primary",
			channel:    user,
			msg:        "request for approving cuebert stop",
		},
	)
}

// parse the input of stopApprover to see if we should stop cuebert
func (b *Bot) stopSubmit(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0]
	approver := ctx.Callback().User.ID
	requester := ctx.Callback().OriginalMessage.Attachments[0].Fallback

	b.log.Trace().
		Str("approver", approver).
		Str("requester", requester).
		Msg("stop submit")

	switch action.Value {
	case ApproveStop:
		_, _, _ = b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
		if approver != requester {
			b.log.Info().Msg("cuebert stop approved")
			b.lifecycle.Stop()
		} else {
			b.stopApprovalOverride()
		}

	case DenyStop:
		_, _, _ = b.bot.SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)

		_, _, err := b.bot.SlackClient().PostMessage(ctx.Callback().User.ID,
			slack.MsgOptionText("Cuebert stop denied :white_check_mark:", false))
		if err != nil {
			b.log.Err(err).Msg("posting message")
		}

		b.log.Info().Msg("cuebert stop denied")
	}
}
