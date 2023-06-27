package main

import (
	"strings"
	"time"

	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

// exceptionHelp invokes the exception modal users can use to request an exception
func (b *Bot) exceptionHelp() {
	definition := &slacker.CommandDefinition{
		Description: "Request an exception",
		Examples:    []string{"request exception"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			b.wantException(botCtx.Event().User, "Would you like to request an exception for updating?")
		},
	}
	b.commands = append(b.commands, Commands{
		usage: "request exception",
		def:   definition,
	})
}

// exceptionRequestDecision handles the decision to approve or deny an exception
func (b *Bot) exceptionRequestDecision(
	s *slacker.Slacker,
	event *socketmode.Event,
	callback *slack.InteractionCallback) {

	if callback.Type != slack.InteractionTypeInteractionMessage {
		return
	}

	action := callback.ActionCallback.AttachmentActions[0].Value

	vals := callback.OriginalMessage.Attachments[0].Fields

	var (
		reason  string
		until   string
		serials string
	)

	for _, val := range vals {
		switch val.Title {
		case "Serial Numbers":
			serials = val.Value
		case "Reason":
			reason = val.Value
		case "Until":
			until = val.Value
		}
	}

	serialSlice := strings.Split(serials, ",")

	ts, err := time.Parse("2006-01-02", until)
	if err != nil {
		b.log.Debug().
			Str("date", until).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	switch action {
	case "approve_exclusion":
		b.log.Info().Msgf("%s - approving exclusion", callback.User.ID)

		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

		s.SocketMode().Ack(*event.Request)

		for _, serial := range serialSlice {
			info, err := b.db.DeviceBySerial(strings.Trim(serial, " "))
			if err != nil {
				b.log.Err(err).Msgf("could not get serial info for %s", serial)
				continue
			}

			email := info[0].User

			slackid, err := b.db.UserByEmail(email)

			if err != nil {
				b.log.Err(err).Msgf("could not get slackid by email %s", email)
				return
			}

			err = b.db.ApproveException(reason, serial, ts)

			if err != nil {
				b.log.Err(err).Msgf("could not approve exclusion for %s", serial)
			}

			_, _, err = s.Client().PostMessage(slackid[0].UserSlackID,
				slack.MsgOptionText("Your request for an exception has been approved :white_check_mark:", false))
			if err != nil {
				b.log.Err(err).Msg("posting exception approval")
			}
		}

	case "deny_exclusion":
		s.SocketMode().Ack(*event.Request)

		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

		_, _, err = s.Client().PostMessage(callback.User.ID,
			slack.MsgOptionText("Your request for an exception has been denied :octagonal_sign:", false))
		if err != nil {
			b.log.Err(err).Msg("posting exception denial")
		}

	default:
		b.log.Debug().Msgf("got an unknown response from exclusion approval")

		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

		s.SocketMode().Ack(*event.Request)
	}

}

// exceptionRequested handles the request for an exception allowing authorized users to approve or deny
func (b *Bot) exceptionRequested(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	if callback.Type != slack.InteractionTypeInteractionMessage {
		return
	}

	action := callback.ActionCallback.AttachmentActions[0]

	switch action.Name {
	case "yes_exception":
		if action.Value == "yes_exception" {
			b.log.Info().Msgf("%s wants to request an exception", callback.User.ID)
		}

		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

		s.SocketMode().Ack(*event.Request)

		devices := b.tables.exceptionSerials(callback.User.ID)

		if len(devices) == 0 {
			_, _, err := s.Client().PostMessage(callback.User.ID,
				slack.MsgOptionText("You have no devices to request an exception for", false))
			if err != nil {
				b.log.Err(err).Msg("posting exception denial")
			}
			return
		}

		b.exceptionRequest(devices, callback.TriggerID)
	case "yes_add_exclusion":
		if action.Value == "yes_add_exclusion" {
			b.log.Debug().Msgf("%s wants to add an exclusion", callback.User.ID)
		}

		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

		s.SocketMode().Ack(*event.Request)

		b.exclusionRequest(callback.TriggerID)
	default:
		_, _, _ = s.Client().DeleteMessage(callback.Channel.ID, callback.MessageTs)

		s.SocketMode().Ack(*event.Request)
	}

}

// wantException asks the user if they want to request an exception and sends the modal
// via exceptionRequest to gather the information.
func (b *Bot) wantException(user, text string) {
	b.modalGateway(&modalGateway{
		text:       text,
		callbackID: "exception_question",
		yesName:    "yes_exception",
		yesText:    "Yes",
		yesValue:   "yes_exception",
		yesStyle:   "primary",
		noName:     "no_exception",
		noText:     "No",
		noValue:    "no_exception",
		noStyle:    "danger",
		channel:    user,
		msg:        "request for exception",
	})
}

// after the user has selected the devices they want to exclude, this function grabs the values to send to the db.
func (b *Bot) exceptionSubmit(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	values := callback.View.State.Values
	dv := values["exception_date_picker"]["datePicker"].SelectedDate
	reason := values["exception_reason"]["exception_input"].Value

	serials := []string{}
	for _, v := range values["user_devices"]["device_box"].SelectedOptions {
		serials = append(serials, v.Text.Text)
	}

	user := callback.User.ID
	b.log.Debug().
		Str("user", user).
		Str("date", dv).
		Str("reason", reason).
		Strs("serials", serials).
		Msg("exception request")

	s.SocketMode().Ack(*event.Request)

	ts, err := time.Parse("2006-01-02", dv)
	if err != nil {
		b.log.Debug().
			Str("date", dv).
			AnErr("error", err).
			Msg("could not compose time string")
	}

	err = b.db.RequestException(callback.User.ID, reason, serials, ts)
	if err != nil {
		b.log.Err(err).Msg("adding exception request to db")
	}

	_, _, err = s.Client().PostMessage(callback.User.ID,
		slack.MsgOptionText("Your request has been submitted :white_check_mark:", false))
	if err != nil {
		b.log.Err(err).Msg("posting ack for exception request")
	}

	b.exclusionApprove(user, reason, dv, serials)
}
