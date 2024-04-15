package shared

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

func (c *Shared) ACK(ctx *slacker.InteractionContext, which string) {
	err := c.caller.MessageAcknowledged(ctx.Callback().User.ID, time.Now().UTC(), which)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not record the first ack time")
	}

	response := fmt.Sprintf("Acknowledged at %v", time.Now().Local().Format(time.RFC1123))

	_, _, err = c.bot.Bot().SlackClient().
		PostMessage(
			ctx.Callback().Channel.ID,
			slack.MsgOptionText(response, false),
			slack.MsgOptionTS(ctx.Callback().MessageTs),
		)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not post message")
	}

	if err = c.bot.Bot().SlackClient().
		AddReaction("white_check_mark", slack.NewRefToMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)); err != nil {
		c.log.Info().AnErr("error", err).Msg("could not add reaction")
	}
}

func (c *Shared) BaseMessage(rp *ReminderPayload, splay int64) (bool, error) {
	c.log.Debug().
		Int64("splay", splay).
		Str("user", rp.UserName).
		Msg("waiting seconds to send message")

	c.log.Debug().Msg("sleep over, sending message")

	attachment := slack.Attachment{
		Title:      rp.Title,
		Text:       rp.FirstMessage,
		CallbackID: c.constFormat(AckIT),
		Color:      "#3AA3E3",
		Actions: []slack.AttachmentAction{
			{
				Name:  "accept",
				Text:  "Acknowledge",
				Type:  "button",
				Value: "ack",
			},
		},
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := c.bot.Bot().SlackClient().PostMessage(rp.UserSlackID, message)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not post message")
		return false, err
	}

	c.log.Trace().
		Str("time", timestamp).
		Str("user", rp.UserName).
		Str("channel", channelID).
		Msg("first message sent")

	return true, nil
}

// finalMessage delivers the final message
func (c *Shared) finalMessage(m *ReminderPayload, splay int64) error {
	c.log.Debug().
		Int64("splay", splay).
		Str("user", m.UserName).
		Msg("sleeping before sending final message")

	time.Sleep(time.Duration(splay) * time.Second)
	c.log.Debug().Msg("sleep over, sending message")

	attachment := slack.Attachment{
		Title:      m.Title,
		Text:       m.Text,
		CallbackID: c.Format(FinalAck),
		Color:      "#3AA3E3",
		Actions: []slack.AttachmentAction{
			{
				Name:  "accept",
				Text:  "Acknowledge",
				Type:  "button",
				Value: "ack",
			},
		},
	}

	message := slack.MsgOptionAttachments(attachment)

	channelID, timestamp, err := c.bot.Bot().SlackClient().PostMessage(m.UserSlackID, message)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not post message")
		return err
	}

	c.log.Trace().
		Str("time", timestamp).
		Str("user", m.UserName).
		Str("channel", channelID).
		Msg("final message sent")

	return nil
}

// MessageTesting acknowledges the request and loads a modal
func (c *Shared) MessageTesting(ctx *slacker.InteractionContext) {
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

func (c *Shared) SendMSG(rp *ReminderPayload, count int) {
	if c.flags.Testing {
		if !helpers.Contains(c.flags.TestUsers, rp.UserSlackID) {
			c.log.Trace().
				Str("user", rp.UserSlackID).
				Bool("testing", c.flags.Testing).
				Strs("test_users", c.flags.TestUsers).
				Msg("not sending message")

			return
		}
	}

	// if !c.shouldSend(rp.TZOffset, rp.UserSlackID) {
	// 	return
	// }

	ex, app := c.caller.IsExcluded(rp.Item)
	if ex {
		c.log.Info().
			Str("serial", rp.Item).
			Str("user", rp.UserName).
			Msg("on the exclusions list. checking approval status.")
		if app {
			c.log.Info().
				Str("serial", rp.Item).
				Bool("approved", app).
				Msg("exclusion approved.")
			return
		}
	}

	switch count {
	case 1:
		err := c.caller.FirstMessageSent(rp.UserServiceID, rp.UserSlackID)
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not record the first message sent")
		}

		// randomize this a scooch so we arent blasting everyone at the same time.
		sn, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not generate random delay")
		}
		go func() {
			ok, err := c.BaseMessage(
				rp,
				sn.Int64(),
			)
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("could not send first message")
			}
			if !ok {
				c.log.Info().Msg("not ok")
			}

			err = c.caller.FirstMessageDelivered(rp.UserServiceID, rp.UserSlackID)
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("could not record the first message delivered")
			}
		}()

	case 2:
		err := c.caller.ReminderMessage(rp)
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not deliver reminder")
		}
	case 3:
		sn, err := rand.Int(rand.Reader, big.NewInt(45))
		if err != nil {
			c.log.Info().AnErr("error", err).Msg("could not generate random delay")
		}

		go func() {
			err = c.finalMessage(rp, sn.Int64())
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("could not send final message")
			}

			err = c.caller.FinalMessageSent(rp.UserServiceID, rp.UserSlackID)
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("could not record the first message sent")
			}
		}()
	default:
		c.log.Trace().Msg("no action at this point")
	}
}
