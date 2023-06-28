package main

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
)

// getSelfInfo returns information about the user when requested.
func (b *Bot) getSelfInfo() {
	definition := &slacker.CommandDefinition{
		Description: "Get info about your account",
		Examples:    []string{"get my info"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			attachments := []slack.Attachment{
				{Color: "blue", AuthorName: "cuebert"},
			}

			user, err := b.db.UserByID(botCtx.Event().User)
			if err != nil {
				b.log.Err(err).Msg("error getting user")
				attachments = append(attachments, slack.Attachment{Color: "red", Text: "error getting user"})
				_ = response.Reply(botCtx.Event().User, slacker.WithAttachments(attachments))
				return
			}

			if !user.Empty() {
				email := user[0].UserEmail
				ui, err := b.tables.user(email)
				if err != nil {
					b.log.Err(err).Msg("error getting user")
				}
				attachments = append(attachments, *ui)
			}

			err = response.Reply(botCtx.Event().User, slacker.WithAttachments(attachments))
			if err != nil {
				b.log.Err(err).Msg("error responding")
			}
		},
	}
	b.commands = append(b.commands, Commands{
		usage: "get my info",
		def:   definition,
	})
}

// getUsersInfo returns information about a user when requested by an admin.
func (b *Bot) getUsersInfo() *slacker.CommandDefinition {
	var reportOpts = []string{"slackid", "email"}
	return &slacker.CommandDefinition{
		Description: "Get info about a user",
		Examples:    []string{"get users info slackid <id>", "get users info email <email>"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			opt := request.Param("opt")
			which := request.Param("input")

			attachments := []slack.Attachment{
				{Color: "blue", AuthorName: "cuebert"},
			}

			var (
				user  bot.BR
				err   error
				email string
			)

			switch strings.ToLower(opt) {
			case "slackid":
				user, err = b.db.UserBySlackID(which)
			case "email":
				which = extractEmails(which)
				user, err = b.db.UserEmail(which)
			default:
				msg := fuzzyMatchNonOpt(opt, reportOpts)
				err := response.Reply(msg)

				if err != nil {
					b.log.Debug().AnErr("sending response", err).Send()
				}
				return
			}

			if err != nil {
				b.log.Debug().AnErr("getting user", err).Send()
			}

			if !user.Empty() {
				email = user[0].UserEmail
				ui, err := b.tables.user(email)
				if err != nil {
					b.log.Err(err).Msg("error getting user")
				}
				attachments = append(attachments, *ui)

				bi := b.tables.bot(user)
				attachments = append(attachments, *bi)

				ei, err := b.tables.exclusions(email)
				if err != nil {
					b.log.Err(err).Msg("error getting exclusions")
				}
				attachments = append(attachments, ei...)

				di, err := b.tables.device(email)
				if err != nil {
					b.log.Err(err).Msg("error getting devices")
				}
				attachments = append(attachments, di...)
			}

			err = response.Reply(botCtx.Event().User, slacker.WithAttachments(attachments))
			if err != nil {
				b.log.Err(err).Msg("error responding")
			}
		},
	}
}

func (t *Tables) user(email string) (*slack.Attachment, error) {
	user, err := t.db.UserByEmail(email)
	if err != nil {
		return nil, err
	}
	var attch slack.Attachment
	for _, u := range user {
		attch = slack.Attachment{
			Title: "User Info",
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
					Value: tzStringer(u.TZOffset),
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

func (t *Tables) bot(user []bot.BotResInfo) *slack.Attachment {
	var attch slack.Attachment
	for _, u := range user {
		attch = slack.Attachment{
			Title: "Bot Results Table",
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

func (t *Tables) exclusions(email string) ([]slack.Attachment, error) {
	user, err := t.db.ExclusionBy().Email(email).Query()
	if err != nil {
		return nil, err
	}

	var attch slack.Attachment
	var attachments []slack.Attachment
	if !user.Empty() {
		for _, u := range user {
			attch = slack.Attachment{
				Title: "Exclusion Info",
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
	attch.Title = "No Exclusions Found"
	attachments = append(attachments, attch)
	return attachments, nil
}

// this makes an assumption the user field is set as email which may not be true in all cases.
func (t *Tables) device(email string) ([]slack.Attachment, error) {
	devices, err := t.db.QueryDeviceBy().User(email).Query()
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

// when the useremail is entered slack will send the email as <mailto:email|email>
func extractEmails(message string) string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindString(message)
	return emails
}

func tzStringer(tz int64) string {
	return strconv.FormatInt(tz, 10)
}
