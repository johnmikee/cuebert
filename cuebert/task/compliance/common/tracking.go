package common

import (
	"strconv"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/slack-go/slack"
)

// Tracking returns the tracking information for the given user as an attachment.
func (c *Common) Tracking(title string, user tracking.TI) *slack.Attachment {
	var attch slack.Attachment
	for _, u := range user {
		attch = slack.Attachment{
			Title: title,
			Fields: []slack.AttachmentField{
				{
					Title: "Slack ID",
					Value: u.UserSlackID,
				},
				{
					Title: "Email",
					Value: u.UserEmail,
				},
				{
					Title: "First Message Acknowledged",
					Value: strconv.FormatBool(u.FirstACK),
				},
				{
					Title: "First Message Acknowledged At",
					Value: u.FirstACKTime.Format(time.RFC3339),
				},
				{
					Title: "First Message Sent",
					Value: strconv.FormatBool(u.FirstMessageSent),
				},
				{
					Title: "First Message Sent At",
					Value: u.FirstMessageSentAt.Format(time.RFC3339),
				},
				{
					Title: "First Message Waiting to Send",
					Value: strconv.FormatBool(u.FirstMessageWaiting),
				},
				{
					Title: "Updated at",
					Value: u.UpdatedAt.Format(time.RFC3339),
				},
			},
		}
	}
	return &attch
}

func (c *Common) CheckNotSent(serviceID string, slackID string) (bool, error) {
	ti, err := c.tables.Tracking().By().Opts(tracking.UserSlackID, slackID).Query()
	if err != nil {
		return true, err
	}
	if ti.Empty() {
		return true, nil
	}

	for _, t := range ti {
		if t.FinalMessageSent { // return false if we have already sent the final message
			return false, nil
		}
	}

	return true, nil
}
