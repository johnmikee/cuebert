package common

import (
	"strconv"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/slack-go/slack"
)

func (c *Common) Bot(title string, user tracking.TI) *slack.Attachment {
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
