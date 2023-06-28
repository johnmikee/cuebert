package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/johnmikee/cuebert/internal/db"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

func (b *Bot) ack(s *slacker.Slacker, event *socketmode.Event, callback *slack.InteractionCallback) {
	if callback.Type != slack.InteractionTypeInteractionMessage {
		return
	}

	action := callback.ActionCallback.AttachmentActions[0]
	if action.Name != "accept" {
		return
	}

	if action.Value == "ack" {
		b.log.Trace().Msgf("%s ackd first message", callback.User.ID)

		err := b.db.ACKACKD(callback.User.ID, time.Now().UTC())
		if err != nil {
			b.log.Err(err).Msg("could not record the first ack time")
		}
	}

	response := fmt.Sprintf("Acknowledged at %v", time.Now().Local().Format(time.RFC1123))

	_, _, err := s.Client().
		PostMessage(
			callback.Channel.ID,
			slack.MsgOptionText(response, false),
			slack.MsgOptionTS(callback.MessageTs),
		)
	if err != nil {
		b.log.Err(err).Msg("could not post message")
	}

	if err = s.Client().AddReaction("white_check_mark",
		slack.NewRefToMessage(callback.Channel.ID, callback.MessageTs)); err != nil {
		b.log.Err(err).Msg("could not add reaction")
	}

	s.SocketMode().Ack(*event.Request)
}

func firstMsg(day time.Time) string {
	return fmt.Sprintf(`
Hello, you are receiving this message because your laptop macOS is out of date.

In order to have continued access to Megacorp Data/Systems (e.g., Gmail, Okta, Zoom) your device must be compliant with our company <https://megacorp.org/securitything|security policies>. 
Our policies *state that your macOS must be up to date* because upgrading your device is crucial for a secure work environment.

If your device continues to stay out of compliance, you will lose access to Megacorp systems at the end of the week.

To upgrade macOS, go to *System Preferences*, and click *Software Update*.

Once you have clicked *Upgrade Now*, the update will begin downloading.  A progress bar will show the status of the download and during this time you can still use your computer as you normally would.

*Your Manager will be engaged if your device is not compliant by %s.*

Post in #the-team-slack-channel if you have any trouble updating your machine.
	`, day.Format("Monday, January 2, 2006"))
}

func (b *Bot) baseMessage(rp *reminderPayload, splay int64) {
	b.log.Debug().Int64("splay", splay).Str("user", rp.userName).Msg("waiting seconds to send message")
	time.Sleep(time.Duration(splay) * time.Second)
	b.log.Debug().Msg("sleep over, sending message")

	attachment := slack.Attachment{
		Title:      fmt.Sprintf("Device: %s", rp.serial),
		Text:       firstMsg(getReminderDay()),
		CallbackID: "ack_it",
		Color:      "#3AA3E3",
		Actions: []slack.AttachmentAction{
			{
				Name:  "accept",
				Text:  "Acknowledge",
				Type:  "button",
				Value: "ack",
			},
		},
		Footer: fmt.Sprintf("Model: %s, OS: %s", rp.model, rp.os),
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := b.bot.Client().PostMessage(rp.userSlackID, message)
	if err != nil {
		b.log.Err(err).Msg("error posting message")
	}

	// the timestamp is returned as an epoch string so we need to convert that
	sec, err := strconv.ParseFloat(timestamp, 64)
	if err != nil {
		b.log.Err(err).Msgf("could not convert the response timestamp %s ", timestamp)
	}
	b.log.Debug().
		Str("serial", rp.serial).
		Str("time", timestamp).
		Str("user", rp.userName).
		Str("channel", channelID).
		Msg("first message sent")

	err = b.db.FirstMessageSent(rp.userSlackID, rp.serial, time.Unix(int64(sec), 0).UTC())
	if err != nil {
		b.log.Info().Msgf("error adding ack: %s", err.Error())
	}
}

func (b *Bot) sendMSG(rp *reminderPayload, count int) {
	// we need to make sure this is sending at a time the user is active.
	// ensure that it is between 9-10am based off the users timezone.
	userLocale := genLocation(rp.tzOffset)
	userTimeNow := time.Now().In(userLocale)

	var (
		start string
		end   string
	)
	if b.cfg.flags.testing {
		start = b.cfg.flags.testingStartTime
		end = b.cfg.flags.testingEndTime
	} else {
		start = "09:00"
		end = "10:00"
	}

	if b.cfg.flags.testing {
		if !helpers.Contains(b.cfg.testUsers, rp.userSlackID) {
			b.log.Trace().
				Str("user", rp.userSlackID).
				Bool("testing", b.cfg.flags.testing).
				Strs("test_users", b.cfg.testUsers).
				Msg("not sending message")

			return
		}
		if !inRange(start, end, userTimeNow) {
			b.log.Debug().
				Str("user", rp.userSlackID).
				Str("user_time", userTimeNow.String()).
				Str("start", start).
				Str("end", end).
				Msg("not in time range")
			return
		}
	} else {
		b.log.Trace().
			Str("user", rp.userSlackID).
			Bool("testing", b.cfg.flags.testing).
			Strs("test_users", b.cfg.testUsers).
			Msg("sending message")
	}

	ex, app := b.tables.isExcluded(rp.serial)
	if ex {
		b.log.Info().
			Str("serial", rp.serial).
			Str("user", rp.userName).
			Msg("on the exclusions list. checking approval status.")
		if app {
			b.log.Info().
				Str("serial", rp.serial).
				Bool("approved", app).
				Msg("exclusion approved.")
			return
		}
	}

	switch count {
	case 1:
		_, err := b.db.FirstMessageWaiting(db.Setter, rp.serial)
		if err != nil {
			b.log.Err(err).Msg("error setting first message waiting")
		}
		// randomize this a scooch so we arent blasting everyone at the same time.
		n, err := rand.Int(rand.Reader, big.NewInt(60))
		if err != nil {
			b.log.Err(err).Msg("could not generate random delay")
		}
		go b.baseMessage(rp, n.Int64())
	case 2:
		err := deliverReminder(
			&reminderInfo{
				deadline: b.cfg.flags.deadline,
				cutoff:   b.cfg.flags.cutoffTime,
				user:     rp.userSlackID,
				serial:   rp.serial,
				version:  b.cfg.flags.requiredVers,
				os:       rp.os,
				text:     "Hi there! Just a gentle reminder to acknowledge the previous message about updating.",
				log:      b.log,
				bot:      b.bot,
			},
		)
		if err != nil {
			b.log.Err(err).Msg("could not deliver reminder")
		}
	default:
		b.log.Trace().Msg("no action at this point")
	}
}
