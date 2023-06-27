package main

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// deliverReminder delivers the reminder to the user
func deliverReminder(ri *reminderInfo) error {
	attachment := slack.Attachment{
		Title:      "Cuebert Update Reminder",
		Text:       ri.text,
		CallbackID: "user_reminder",
		Color:      "#3AA3E3",
		Fields: []slack.AttachmentField{
			{
				Title: "Required Version",
				Value: ri.version,
			},
			{
				Title: "Update Deadline",
				Value: ri.deadline + " " + ri.cutoff,
			},
			{
				Title: "Current Version",
				Value: ri.os,
			},
			{
				Title: "Serial Number",
				Value: ri.serial,
			},
		},
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := ri.bot.Client().PostMessage(ri.user, message)
	if err != nil {
		ri.log.Err(err).Msg("posting message")
		return err
	}

	ri.log.Debug().
		Str("user", ri.user).
		Str("serial", ri.serial).
		Str("channel", channelID).
		Str("timestamp", timestamp).
		Str("table", "bot_results").
		Bool("sent", true).
		Msg("delivered reminder")
	return nil
}

// deliverReminder sends a reminder to the user based on their input
func (b *Bot) deliverReminder(count int, x *bot.BotResInfo) {
	dev, err := b.db.DeviceBySerial(x.SerialNumber)
	if err != nil {
		b.log.Info().
			AnErr("getting devices", err).
			Str("user", x.SlackID).
			Str("serial", x.SerialNumber).
			Send()
		return
	}

	b.sendMSG(
		&reminderPayload{
			userSlackID:    x.SlackID,
			managerSlackID: x.ManagerSlackID,
			userName:       x.FullName,
			serial:         x.SerialNumber,
			model:          dev[0].Model,
			os:             dev[0].OSVersion,
			tzOffset:       x.TZOffset,
		},
		count,
	)
}

// reminderHelp adds the reminder command to the bot
func (b *Bot) reminderHelp() {
	definition := &slacker.CommandDefinition{
		Description: "Request a reminder to update",
		Examples:    []string{"request reminder"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			b.wantAReminder(botCtx.Event().User)
		},
	}
	b.commands = append(b.commands, Commands{
		usage: "request reminder",
		def:   definition,
	})
}

// reminderRequested is the callback for the reminder button
func (b *Bot) reminderRequested(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	if callback.Type != slack.InteractionTypeInteractionMessage {
		return
	}

	action := callback.ActionCallback.AttachmentActions[0]
	if action.Name != "remind_me" {
		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

		s.SocketMode().Ack(*event.Request)
		return
	}

	if action.Value == "remind_me" {
		b.log.Info().Msgf("%s wants a reminder to update", callback.User.ID)

		err := b.db.ACKACKD(callback.User.ID, time.Now().UTC())
		if err != nil {
			b.log.Err(err).Msg("could not record the first ack time")
		}
	}

	_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

	s.SocketMode().Ack(*event.Request)

	b.reminderPicker(callback.TriggerID, "Please enter a time to be reminded")
}

// reminderSubmit is the callback for the reminder picker modal
func (b *Bot) reminderSubmit(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	values := callback.View.State.Values

	dv := values["date_picker"]["datePicker"]
	tv := values["time_picker"]["timePicker"]

	b.log.Trace().Msgf("%s is setting a reminder for updating: %s %s", callback.User.ID, dv.SelectedDate, tv.SelectedTime)

	// get the users offset
	ui, err := b.db.UserByID(callback.User.ID)
	if err != nil {
		b.log.Err(err).Msg("could not get user")
	}
	offset := ui[0].TZOffset
	// validate this is not in the past.
	if !futureDate(dv.SelectedDate, tv.SelectedTime, offset) {
		b.log.Debug().Msgf("%s set a date in the past.", callback.User.ID)

		s.SocketMode().Ack(*event.Request)

		update := fmt.Sprintf(
			"Sorry %s %s already happened..\nPlease set a date in the future. :clock1:",
			dv.SelectedDate,
			tv.SelectedTime,
		)
		_, _, err := s.Client().PostMessage(callback.User.ID, slack.MsgOptionText(update, false))

		if err != nil {
			b.log.Err(err).Msg("posting time fix message")
		}
		b.wantAReminder(callback.User.ID)
		return
	}

	b.tables.updateReminderTime(dv.SelectedDate, tv.SelectedTime, callback.User.ID)

	s.SocketMode().Ack(*event.Request)

	update := fmt.Sprintf("Your reminder has been set for %s %s :clock1:", dv.SelectedDate, tv.SelectedTime)

	_, _, err = s.Client().PostMessage(callback.User.ID, slack.MsgOptionText(update, false))
	if err != nil {
		b.log.Err(err).Msg("posting time fix message")
	}
}

type reminderInfo struct {
	deadline string
	cutoff   string
	user     string
	serial   string
	version  string
	os       string
	text     string
	log      logger.Logger
	bot      *slacker.Slacker
}

// scheduleReminder will execute the scheduled reminder set by the user.
func (b *Bot) scheduleReminder(t time.Duration, ri *reminderInfo) {
	sent, err := b.db.ReminderSentCheck(ri.user)

	if err != nil {
		b.log.Info().
			Str("user", ri.user).
			AnErr("checking if reminder has been sent", err).
			Send()
		return
	}

	if sent {
		b.log.Debug().
			Str("user", ri.user).
			Bool("sent", sent).
			Msg("not sending reminder, already sent")
		return
	}

	b.log.Debug().
		Str("user", ri.user).
		Float64("sleeping_seconds", t.Seconds()).
		Msg("waiting to send reminder")

	time.Sleep(t)

	b.log.Trace().
		Str("user", ri.user).
		Str("serial", ri.serial).
		Msg("sending requested reminder")

	err = deliverReminder(ri)

	if err != nil {
		b.log.Info().
			Str("user", ri.user).
			AnErr("delivering reminder", err).
			Send()
		return
	}

	err = b.db.ReminderSent(true, ri.serial)

	if err != nil {
		b.log.Info().
			Str("user", ri.user).
			Str("serial", ri.serial).
			Str("table", "bot_results").
			AnErr("updating table", err).
			Send()
		return
	}

	b.log.Trace().
		Str("user", ri.user).
		Str("serial", ri.serial).
		Str("table", "bot_results").
		Bool("updated", true).
		Msg("notification reminder sent")
}

// wantAReminder provides access to the reminder modal.
func (b *Bot) wantAReminder(user string) {
	b.modalGateway(&modalGateway{
		text:       "Would you like a reminder to update your device?",
		callbackID: "remind_me_question",
		yesName:    "remind_me",
		yesText:    "Yes",
		yesValue:   "remind_me",
		yesStyle:   "primary",
		noName:     "dont_remind_me",
		noText:     "No",
		noValue:    "dont_remind_me",
		noStyle:    "danger",
		channel:    user,
		msg:        "reminder requested",
	})
}
