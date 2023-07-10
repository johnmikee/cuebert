package bot

import (
	"strconv"
	"time"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// getSelfInfo returns information about the user when requested.
func (b *Bot) getSelfInfo() {
	definition := &slacker.CommandDefinition{
		Command:     "get my info",
		Description: "Get info about your account",
		Examples:    []string{"get my info"},
		Handler: func(ctx *slacker.CommandContext) {
			attachments := []slack.Attachment{
				{Color: "blue", AuthorName: "cuebert"},
			}
			user, err := b.tables.UserByID(ctx.Event().UserID)
			if err != nil {
				b.log.Err(err).Msg("error getting user")
				attachments = append(attachments, slack.Attachment{Color: "red", Text: "error getting user"})

				_, err := ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
				if err != nil {
					b.log.Err(err).Msg("error responding")
				}
				return
			}

			if !user.Empty() {
				email := user[0].UserEmail
				ui, err := b.User("User Info", email)
				if err != nil {
					b.log.Err(err).Msg("error getting user")
				}
				attachments = append(attachments, *ui)
			}

			_, err = ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
			if err != nil {
				b.log.Err(err).Msg("error responding")
			}
		},
	}
	b.bot.AddCommand(definition)
}

func (b *Bot) User(title, email string) (*slack.Attachment, error) {
	user, err := b.tables.UserByEmail(email)
	if err != nil {
		return nil, err
	}
	var attch slack.Attachment
	for _, u := range user {
		attch = slack.Attachment{
			Title: title,
			Fields: []slack.AttachmentField{
				{
					Title: "MDM ID",
					Value: u.MDMID,
				},
				{
					Title: "User Name",
					Value: u.UserLongName,
				},
				{
					Title: "User Email",
					Value: u.UserEmail,
				},
				{
					Title: "User Slack ID",
					Value: u.UserSlackID,
				},
				{
					Title: "User Timezone",
					Value: helpers.TZStringer(u.TZOffset),
				},
				{
					Title: "Created at",
					Value: u.CreatedAt.Format(time.RFC3339),
				},
				{
					Title: "Updated at",
					Value: u.UpdatedAt.Format(time.RFC3339),
				},
			},
		}

	}

	return &attch, nil
}

func (b *Bot) Bot(title string, user bot.BR) *slack.Attachment {
	var attch slack.Attachment
	for _, u := range user {
		attch = slack.Attachment{
			Title: title,
			Fields: []slack.AttachmentField{
				{
					Title: "Slack ID",
					Value: u.SlackID,
				},
				{
					Title: "Email",
					Value: u.UserEmail,
				},
				{
					Title: "Manager Slack ID",
					Value: u.ManagerSlackID,
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
					Title: "Created at",
					Value: u.CreatedAt.Format(time.RFC3339),
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

func (b *Bot) Exclusions(email string) ([]slack.Attachment, error) {
	user, err := b.tables.ExclusionBy().Email(email).Query()
	if err != nil {
		return nil, err
	}

	var attch slack.Attachment
	var attachments []slack.Attachment
	if !user.Empty() {
		for _, u := range user {
			attch = slack.Attachment{
				Title: "exception Info",
				Fields: []slack.AttachmentField{
					{
						Title: "Approved",
						Value: strconv.FormatBool(u.Approved),
					},
					{
						Title: "Created At",
						Value: u.CreatedAt.Format(time.RFC3339),
					},
					{
						Title: "Serial Number",
						Value: u.SerialNumber,
					},
					{
						Title: "Reason",
						Value: u.Reason,
					},
					{
						Title: "Updated At",
						Value: u.UpdatedAt.Format(time.RFC3339),
					},
					{
						Title: "User Email",
						Value: u.UserEmail,
					},
					{
						Title: "Until",
						Value: u.Until.Format(time.RFC3339),
					},
				},
			}
			attachments = append(attachments, attch)
		}
		return attachments, nil
	}
	attch.Title = "No exclusions Found"
	attachments = append(attachments, attch)
	return attachments, nil
}

// this makes an assumption the user field is set as email which may not be true in all cases.
func (b *Bot) Device(email string) ([]slack.Attachment, error) {
	devices, err := b.tables.QueryDeviceBy().User(email).Query()
	if err != nil {
		return nil, err
	}

	var attch slack.Attachment
	var attachments []slack.Attachment
	if !devices.Empty() {
		for _, d := range devices {
			attch = slack.Attachment{
				Title: "Devices Info",
				Fields: []slack.AttachmentField{
					{
						Title: "Device ID",
						Value: d.DeviceID,
					},
					{
						Title: "Device Name",
						Value: d.DeviceName,
					},
					{
						Title: "Last Check In",
						Value: d.LastCheckIn.Format(time.RFC3339),
					},
					{
						Title: "Model",
						Value: d.Model,
					},
					{
						Title: "OS Version",
						Value: d.OSVersion,
					},
					{
						Title: "Platform",
						Value: d.Platform,
					},
					{
						Title: "Serial Number",
						Value: d.SerialNumber,
					},
					{
						Title: "User",
						Value: d.User,
					},
					{
						Title: "User MDM ID",
						Value: d.UserMDMID,
					},
					{
						Title: "Created At",
						Value: d.CreatedAt.Format(time.RFC3339),
					},
					{
						Title: "Updated At",
						Value: d.UpdatedAt.Format(time.RFC3339),
					},
				},
			}
		}
		attachments = append(attachments, attch)
		return attachments, nil
	}
	attch.Title = "No Devices Found"
	attachments = append(attachments, attch)
	return attachments, nil
}
