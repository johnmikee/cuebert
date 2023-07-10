package bot

import (
	"fmt"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// reminderSubmit is the callback for the reminder picker modal
func (b *Bot) reminderSubmit(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values

	dv := values[DatePicker][DatePicker]
	tv := values[TimePicker][TimePicker]

	b.log.Trace().Msgf("%s is setting a reminder for updating: %s %s", ctx.Callback().User.ID, dv.SelectedDate, tv.SelectedTime)

	// get the users offset
	ui, err := b.tables.UserByID(ctx.Callback().User.ID)
	if err != nil {
		b.log.Err(err).Msg("could not get user")
	}
	offset := ui[0].TZOffset
	// validate this is not in the past.
	if !helpers.FutureDate(dv.SelectedDate, tv.SelectedTime, offset) {
		b.log.Debug().Msgf("%s set a date in the past.", ctx.Callback().User.ID)

		update := fmt.Sprintf(
			"Sorry %s %s already happened..\nPlease set a date in the future. :clock1:",
			dv.SelectedDate,
			tv.SelectedTime,
		)
		_, _, err := b.bot.SlackClient().PostMessage(ctx.Callback().User.ID, slack.MsgOptionText(update, false))

		if err != nil {
			b.log.Err(err).Msg("posting time fix message")
		}
		return
	}

	b.tables.UpdateReminderTime(dv.SelectedDate, tv.SelectedTime, ctx.Callback().User.ID)

	update := fmt.Sprintf("Your reminder has been set for %s %s :clock1:", dv.SelectedDate, tv.SelectedTime)

	_, _, err = b.bot.SlackClient().PostMessage(ctx.Callback().User.ID, slack.MsgOptionText(update, false))
	if err != nil {
		b.log.Err(err).Msg("posting time fix message")
	}
}

// SendReminder sends a reminder to the user based on their input
func (b *Bot) SendReminder(count int, x *bot.Info) {
	dev, err := b.tables.DeviceBySerial(x.SerialNumber)
	if err != nil {
		b.log.Info().
			AnErr("getting devices", err).
			Str("user", x.SlackID).
			Str("serial", x.SerialNumber).
			Send()
		return
	}

	b.sendMSG(
		&ReminderPayload{
			UserSlackID: x.SlackID,
			UserName:    x.FullName,
			Serial:      x.SerialNumber,
			Model:       dev[0].Model,
			OS:          dev[0].OSVersion,
			TZOffset:    x.TZOffset,
		},
		count,
	)
}
