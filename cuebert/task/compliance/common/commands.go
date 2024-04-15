package common

import (
	"fmt"

	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/version"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// RegisterHelp registers the help command with the slackbot.
func (c *Common) RegisterHelp() {
	definition := &slacker.CommandDefinition{
		Command:     "help",
		Description: "help!",
		Middlewares: nil,
		Handler: func(ctx *slacker.CommandContext) {
			header := slack.NewTextBlockObject(
				slack.MarkdownType,
				"Hi there :partyblob: here are some things I can help with:\n",
				false,
				false,
			)
			// user commands
			user := slack.NewTextBlockObject(slack.MarkdownType, "*User Info*\n", false, false)
			ui := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "`get my info`\n", false, false),
			}
			ue := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "`check my compliance`\n", false, false),
			}
			// reminders
			reminders := slack.NewTextBlockObject(slack.MarkdownType, "*Reminder*\n", false, false)
			reminder := []*slack.TextBlockObject{
				slack.NewTextBlockObject(
					slack.MarkdownType,
					"Set a reminder to update:\n`request reminder`\n",
					false,
					false,
				),
			}

			// exclusions
			exclusions := slack.NewTextBlockObject(slack.MarkdownType, "*Exclusions*\n", false, false)
			requestex := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Request Exclusion:\n`request exclusion`\n", false, false),
			}
			requestadd := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Add Exclusion:\n`add exclusion`\n", false, false),
			}

			// reports
			reports := slack.NewTextBlockObject(slack.MarkdownType, "*Reports*\n", false, false)
			reportGet := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Get Report:\n`get report`\n", false, false),
			}

			// user commands for admins
			userAdmin := slack.NewTextBlockObject(slack.MarkdownType, "*User Commands*\n", false, false)
			getUser := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Get User Info:\n`get users info slackid <id>`\n", false, false),
			}
			updateUser := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Update User Info:\n`update user interactive`\n", false, false),
			}

			// lifecycle
			lifecycle := slack.NewTextBlockObject(slack.MarkdownType, "*Lifecycle*\n", false, false)
			start := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Start:\n`start`\n", false, false),
			}
			status := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Status:\n`status`\n", false, false),
			}
			stop := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Stop:\n`stop`\n", false, false),
			}
			update := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Update:\n`update`\n", false, false),
			}
			// more
			more := slack.NewTextBlockObject(slack.MarkdownType, "*More*\n", false, false)

			docsButton := slack.ButtonBlockElement{
				Type: slack.METButton,
				Text: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: ":page_with_curl: Documentation",
				},
				ActionID: "docs_btn",
				Value:    "docs_btn",
				URL:      c.flags.HelpDocsURL,
			}

			helpButton := slack.ButtonBlockElement{
				Type: slack.METButton,
				Text: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: ":speech_balloon: Contact Support",
				},
				ActionID: "help_btn",
				Value:    "help_btn",
				URL:      c.flags.HelpTicketURL,
			}

			versionButton := slack.ButtonBlockElement{
				Type: slack.METButton,
				Text: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: ":information_source: Version",
				},
				ActionID: "version_btn",
				Value:    "version_btn_val",
				URL:      c.flags.HelpRepoURL,
				Confirm: slack.NewConfirmationBlockObject(
					&slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "Version Information",
					},
					&slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: fmt.Sprintf("Version: %s\nBranch: %s\nRevision: %s\nGo Version: %s\nBuild Date: %s\nBuild User: %s\n",
							version.Version().Version,
							version.Version().Branch,
							version.Version().Revision,
							version.Version().GoVersion,
							version.Version().BuildDate,
							version.Version().BuildUser),
					},
					&slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "Open Repo",
					},
					&slack.TextBlockObject{
						Type: slack.PlainTextType,
						Text: "Close",
					},
				),
			}
			// build the base blocks
			blocks := []slack.Block{
				slack.NewSectionBlock(header, nil, nil),
				slack.NewDividerBlock(),
				slack.NewSectionBlock(user, nil, nil),
				slack.NewSectionBlock(nil, ui, nil),
				slack.NewSectionBlock(nil, ue, nil),
				slack.NewDividerBlock(),
				slack.NewSectionBlock(reminders, nil, nil),
				slack.NewSectionBlock(nil, reminder, nil),
				slack.NewDividerBlock(),
				slack.NewSectionBlock(exclusions, nil, nil),
				slack.NewSectionBlock(nil, requestex, nil),
			}

			// if the user is in the auth list add the exclusions command
			if helpers.Contains(c.flags.AuthUsers, ctx.Event().UserID) {
				blocks = append(blocks,
					slack.NewSectionBlock(nil, requestadd, nil),
					slack.NewDividerBlock(),
					slack.NewSectionBlock(reports, nil, nil),
					slack.NewSectionBlock(nil, reportGet, nil),
					slack.NewDividerBlock(),
					slack.NewSectionBlock(userAdmin, nil, nil),
					slack.NewSectionBlock(nil, getUser, nil),
					slack.NewSectionBlock(nil, updateUser, nil),
					slack.NewDividerBlock(),
					slack.NewSectionBlock(lifecycle, nil, nil),
					slack.NewSectionBlock(nil, status, nil),
					slack.NewSectionBlock(nil, start, nil),
					slack.NewSectionBlock(nil, stop, nil),
					slack.NewSectionBlock(nil, update, nil),
					slack.NewDividerBlock(),
				)
			}

			// add the rest
			blocks = append(blocks,
				slack.NewDividerBlock(),
				slack.NewSectionBlock(more, nil, nil),
				slack.NewActionBlock("", docsButton, helpButton, versionButton),
				slack.NewDividerBlock(),
			)

			_, err := ctx.Response().ReplyBlocks(blocks)
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("posting help message")
			}
		},
	}
	c.bot.Bot().Help(definition)
}

// GetSelfInfo returns information about the user when requested.
func (c *Common) GetSelfInfo() {
	definition := &slacker.CommandDefinition{
		Command:     "get my info",
		Description: "Get info about your account",
		Examples:    []string{"get my info"},
		Handler: func(ctx *slacker.CommandContext) {
			attachments := []slack.Attachment{
				{Color: "blue", AuthorName: "cuebert"},
			}
			user, err := c.tables.User().By().
				Opts(users.UserSlackID, ctx.Event().UserID).Query()
			c.log.Info().Str("slackid", ctx.Event().UserID).Msg("getting user")
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("error getting user")
				attachments = append(attachments, slack.Attachment{Color: "red", Text: "error getting user"})

				_, err := ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
				if err != nil {
					c.log.Err(err).Msg("error responding")
				}
				return
			}

			if !user.Empty() {
				email := user[0].UserEmail
				ui, err := c.User("User Info", email)
				if err != nil {
					c.log.Info().AnErr("error", err).Msg("error getting user")
				}
				attachments = append(attachments, *ui)
			}

			_, err = ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
			if err != nil {
				c.log.Info().AnErr("error", err).Msg("error responding")
			}
		},
	}
	c.bot.Bot().AddCommand(definition)
}

// User returns a slack attachment with user information.
func (c *Common) User(title, email string) (*slack.Attachment, error) {
	user, err := c.tables.User().By().Opts(users.UserEmail, email).Query()
	if err != nil {
		return nil, err
	}

	if user.Empty() {
		return nil, fmt.Errorf("no user found with email %s", email)
	}

	var attch slack.Attachment
	for _, u := range user {
		attch = slack.Attachment{
			Title: title,
			Fields: []slack.AttachmentField{
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
			},
		}
	}

	return &attch, nil
}
