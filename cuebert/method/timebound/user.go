package timebound

import (
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// UserUpdateModalSubmit implements method.Actions.
func (t *TimeBound) UserUpdateModalSubmit(base *bot.Info, values map[string]map[string]slack.BlockAction) {
	t.log.Trace().Msg("not implemented")
}

// UserUpdate implements method.Actions.
func (t *TimeBound) UserUpdate(ctx *slacker.InteractionContext) {
	t.log.Trace().Msg("not implemented")
}
