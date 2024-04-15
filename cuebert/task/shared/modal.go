package shared

import (
	"github.com/shomali11/slacker/v2"
)

// ModalSubmit implements task.Rabbit.
func (c *Shared) ModalSubmit(ctx *slacker.InteractionContext) {
	switch ctx.Callback().View.CallbackID {
	case c.constFormat(AdminExclusionModal):
		// c.impl.ExclusionSubmit(ctx)
	case c.constFormat(ExclusionModal):
		c.caller.ExclusionSubmit(ctx)
	case c.constFormat(ReminderPicker):
		c.ReminderSubmit(ctx)
	case c.constFormat(UpdateUserSelector):
		// maybe this is fixed now?
		// it was  b.userBRUpdate(s, event, callback, TriggerID)
		c.log.Trace().Interface("callback", ctx.Callback()).Msg("update user selector")
		// c.impl.UserUpdate(ctx)
	case c.constFormat(UserUpdateModal):
		// c.impl.UserUpdateSubmit(ctx)
	case c.constFormat(Start), c.constFormat(Reload):
		c.LoadInput(ctx, ctx.Callback().View.CallbackID)
	case c.constFormat(UpdateBeforeLaunch):
		c.Launch(ctx)
	default:
		c.log.Debug().Str("callback", ctx.Callback().View.CallbackID).Msg("not something we are handling.")
		return
	}
}
