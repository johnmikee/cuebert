package shared

import (
	"strings"

	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

func (c *Shared) InteractionHandler(ctx *slacker.InteractionContext) {
	c.log.Info().Interface("callback", ctx.Callback()).Msg("interaction handler")
	switch ctx.Callback().CallbackID {
	case c.constFormat(AckIT):
		c.ACK(ctx, "first")
	case c.constFormat(FinalAck):
		c.ACK(ctx, "final")
	case c.constFormat(CheckComplianceStatus):
		c.ComplianceChecker(ctx)
	case c.constFormat(RemindMeQuestion):
		c.ReminderRequested(ctx)
	case c.constFormat(MessageRequestYes):
		c.MessageTesting(ctx)
	case c.constFormat(StopCuebertRequest):
		c.StopApprover(ctx)
	case c.constFormat(StopCuebertApproval):
		c.StopSubmit(ctx)
	case c.constFormat(StopCuebertOverride):
		c.OverrideSubmit(ctx)
	case c.constFormat(Start), c.constFormat(Reload), c.constFormat(UpdateUserQuestion):
		c.InteractiveHelper(ctx)
	case c.constFormat(UpdateBeforeLaunch):
		c.Launch(ctx)
	case c.constFormat(ExclusionApprover):
		c.caller.ExclusionRequestDecision(ctx)
	case c.constFormat(ExclusionQuestion), c.constFormat(ExclusionAddQuestion):
		c.ExclusionRequested(ctx)
	case c.constFormat(AdminExclusionModal):
		c.caller.ExclusionSubmit(ctx)
	case c.constFormat(ExclusionModal):
		c.caller.UserExclusionSubmit(ctx)
	case c.constFormat(UpdateUserSelector):
		c.caller.UserUpdate(ctx)
	case c.constFormat(UserUpdateModal):
		c.caller.UserUpdateSubmit(ctx)
	case "":
		switch ctx.Callback().Type {
		case slack.InteractionTypeViewSubmission:
			// these modals dont have a callback id for reasons that are
			// escaping me. pass it to the modalSubmit for further validation
			c.ModalSubmit(ctx)
		default:
			c.log.Trace().
				Str("callback", ctx.Callback().CallbackID).
				Msg("not an interaction we are handling")
		}
	}
}

func (c *Shared) InteractiveHelper(ctx *slacker.InteractionContext) {
	c.log.Info().Interface("callback", ctx.Callback()).Msg("interactive helper")

	action := ctx.Callback().ActionCallback.AttachmentActions[0]
	c.log.Info().Interface("action", action).Msg("action")
	switch action.Name {
	case c.constFormat(Accept):
		c.ACK(ctx, "first")
	case c.constFormat(UpdateUserYes):
		c.UserSelector(ctx.Callback().TriggerID, ctx.Callback().Channel.ID)
	case c.constFormat(UpdateUserSubmit):
		c.caller.UserUpdate(ctx)
	case c.constFormat(UpdateConfigYes), c.constFormat(Reload):
		c.LoadProgram(
			&Loader{
				TriggerID: ctx.Callback().TriggerID,
				LoadType:  ctx.Callback().CallbackID,
				Tables:    Tables(c.flags.TaskImplementation, TaskType(strings.ToLower(c.flags.Task))),
			},
		)
	}
	_, _, _ = c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
}
