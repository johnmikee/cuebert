package bot

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/johnmikee/cuebert/cuebert/tables"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

func (b *Bot) ack(ctx *slacker.InteractionContext) {
	b.log.Trace().Msgf("%s ackd first message", ctx.Callback().User.ID)

	err := b.tables.ACKACKD(ctx.Callback().User.ID, time.Now().UTC())
	if err != nil {
		b.log.Err(err).Msg("could not record the first ack time")
	}

	response := fmt.Sprintf("Acknowledged at %v", time.Now().Local().Format(time.RFC1123))

	_, _, err = b.bot.SlackClient().
		PostMessage(
			ctx.Callback().Channel.ID,
			slack.MsgOptionText(response, false),
			slack.MsgOptionTS(ctx.Callback().MessageTs),
		)
	if err != nil {
		b.log.Err(err).Msg("could not post message")
	}

	if err = b.bot.SlackClient().AddReaction("white_check_mark",
		slack.NewRefToMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)); err != nil {
		b.log.Err(err).Msg("could not add reaction")
	}

}

func (b *Bot) BaseMessage(rp *ReminderPayload, splay int64) {
	b.log.Debug().
		Int64("splay", splay).
		Str("user", rp.UserName).
		Msg("waiting seconds to send message")

	time.Sleep(time.Duration(splay) * time.Second)
	b.log.Debug().Msg("sleep over, sending message")

	attachment := slack.Attachment{
		Title:      fmt.Sprintf("Device: %s", rp.Serial),
		Text:       b.method.FirstMessage(),
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
		Footer: fmt.Sprintf("Model: %s, OS: %s", rp.Model, rp.OS),
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := b.bot.SlackClient().PostMessage(rp.UserSlackID, message)
	if err != nil {
		b.log.Err(err).Msg("error posting message")
	}

	// the timestamp is returned as an epoch string so we need to convert that
	sec, err := strconv.ParseFloat(timestamp, 64)
	if err != nil {
		b.log.Err(err).Msgf("could not convert the response timestamp %s ", timestamp)
	}
	b.log.Debug().
		Str("serial", rp.Serial).
		Str("time", timestamp).
		Str("user", rp.UserName).
		Str("channel", channelID).
		Msg("first message sent")

	err = b.tables.FirstMessageSent(rp.UserSlackID, rp.Serial, time.Unix(int64(sec), 0).UTC())
	if err != nil {
		b.log.Info().Msgf("error adding ack: %s", err.Error())
	}
}

func (b *Bot) sendMSG(rp *ReminderPayload, count int) {
	// we need to make sure this is sending at a time the user is active.
	// ensure that it is between 9-10am based off the users timezone.
	userLocale := helpers.GenLocation(rp.TZOffset)
	userTimeNow := time.Now().In(userLocale)

	var (
		start string
		end   string
	)
	if b.cfg.testing {
		start = b.cfg.testingStartTime
		end = b.cfg.testingEndTime
	} else {
		start = "09:00"
		end = "10:00"
	}

	if b.cfg.testing {
		if !helpers.Contains(b.cfg.testUsers, rp.UserSlackID) {
			b.log.Trace().
				Str("user", rp.UserSlackID).
				Bool("testing", b.cfg.testing).
				Strs("test_users", b.cfg.testUsers).
				Msg("not sending message")

			return
		}
		if !helpers.InRange(start, end, userTimeNow) {
			b.log.Debug().
				Str("user", rp.UserSlackID).
				Str("user_time", userTimeNow.String()).
				Str("start", start).
				Str("end", end).
				Msg("not in time range")
			return
		}
	} else {
		b.log.Trace().
			Str("user", rp.UserSlackID).
			Bool("testing", b.cfg.testing).
			Strs("test_users", b.cfg.testUsers).
			Msg("sending message")
	}

	ex, app := b.tables.IsExcluded(rp.Serial)
	if ex {
		b.log.Info().
			Str("serial", rp.Serial).
			Str("user", rp.UserName).
			Msg("on the exclusions list. checking approval status.")
		if app {
			b.log.Info().
				Str("serial", rp.Serial).
				Bool("approved", app).
				Msg("exclusion approved.")
			return
		}
	}

	switch count {
	case 1:
		_, err := b.tables.FirstMessageWaiting(tables.Setter, rp.Serial)
		if err != nil {
			b.log.Err(err).Msg("error setting first message waiting")
		}
		// randomize this a scooch so we arent blasting everyone at the same time.
		n, err := rand.Int(rand.Reader, big.NewInt(60))
		if err != nil {
			b.log.Err(err).Msg("could not generate random delay")
		}
		go b.BaseMessage(rp, n.Int64())
	case 2:
		err := b.deliverReminder(
			&ReminderInfo{
				Deadline: b.cfg.deadline,
				Cutoff:   b.cfg.cutoffTime,
				User:     rp.UserSlackID,
				Serial:   rp.Serial,
				Version:  b.cfg.requiredVers,
				OS:       rp.OS,
				Text:     "Hi there! Just a gentle reminder to acknowledge the previous message about updating.",
			},
		)
		if err != nil {
			b.log.Err(err).Msg("could not deliver reminder")
		}
	case 3:
		err := b.method.ReminderMessage(rp)
		if err != nil {
			b.log.Err(err).Msg("could not send reminder message")
		}
	default:
		b.log.Trace().Msg("no action at this point")
	}
}
