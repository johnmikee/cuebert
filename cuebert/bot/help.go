package bot

import (
	"fmt"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/version"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// add the help/docs links as input pulled from the config
func (b *Bot) help() *slacker.CommandDefinition {
	return &slacker.CommandDefinition{
		Command:     "help",
		Description: "help!",
		Handler: func(ctx *slacker.CommandContext) {
			header := slack.NewTextBlockObject(
				slack.MarkdownType,
				"Hi there :partyblob: here are some things I can help with:\n",
				false,
				false,
			)
			// device commands
			device := slack.NewTextBlockObject(slack.MarkdownType, "*Device Info*\n", false, false)
			model := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Model:\n`cuebert get device model`\n", false, false),
			}
			serial := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Serial Number(s):\n`cuebert get device serial`\n", false, false),
			}
			os := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Operating System:\n`cuebert get device os`\n", false, false),
			}
			hostname := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Hostname:\n`cuebert get device hostname`\n", false, false),
			}

			// user commands
			user := slack.NewTextBlockObject(slack.MarkdownType, "*User Info*\n", false, false)
			ui := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "`cuebert get my info`\n", false, false),
			}

			// reminders
			reminders := slack.NewTextBlockObject(slack.MarkdownType, "*Reminder*\n", false, false)
			reminder := []*slack.TextBlockObject{
				slack.NewTextBlockObject(
					slack.MarkdownType,
					"Set a reminder to update:\n`cuebert request reminder`\n",
					false,
					false,
				),
			}

			// exclusions
			exclusions := slack.NewTextBlockObject(slack.MarkdownType, "*Exclusions*\n", false, false)
			requestex := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Request Exclusion:\n`cuebert request exclusion`\n", false, false),
			}
			requestadd := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Add Exclusion:\n`cuebert add exclusion`\n", false, false),
			}

			// reports
			reports := slack.NewTextBlockObject(slack.MarkdownType, "*Reports*\n", false, false)
			reportGet := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Get Report:\n`cuebert get report`\n", false, false),
			}

			// user commands for admins
			userAdmin := slack.NewTextBlockObject(slack.MarkdownType, "*User Commands*\n", false, false)
			getUser := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Get User Info:\n`cuebert get users info slackid <id>`\n", false, false),
			}
			updateUser := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Update User Info:\n`cuebert update user interactive`\n", false, false),
			}

			// lifecycle
			lifecycle := slack.NewTextBlockObject(slack.MarkdownType, "*Lifecycle*\n", false, false)
			start := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Start:\n`start cuebert`\n", false, false),
			}
			stop := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Stop:\n`stop cuebert`\n", false, false),
			}
			update := []*slack.TextBlockObject{
				slack.NewTextBlockObject(slack.MarkdownType, "Update:\n`update cuebert`\n", false, false),
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
				URL:      b.cfg.helpDocsURL,
			}

			helpButton := slack.ButtonBlockElement{
				Type: slack.METButton,
				Text: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: ":speech_balloon: Contact Support",
				},
				ActionID: "help_btn",
				Value:    "help_btn",
				URL:      b.cfg.helpTicketURL,
			}

			versionButton := slack.ButtonBlockElement{
				Type: slack.METButton,
				Text: &slack.TextBlockObject{
					Type: slack.PlainTextType,
					Text: ":information_source: Version",
				},
				ActionID: "version_btn",
				Value:    "version_btn_val",
				URL:      b.cfg.helpRepoURL,
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
				slack.NewSectionBlock(device, nil, nil),
				slack.NewSectionBlock(nil, serial, nil),
				slack.NewSectionBlock(nil, model, nil),
				slack.NewSectionBlock(nil, os, nil),
				slack.NewSectionBlock(nil, hostname, nil),
				slack.NewDividerBlock(),
				slack.NewSectionBlock(user, nil, nil),
				slack.NewSectionBlock(nil, ui, nil),
				slack.NewDividerBlock(),
				slack.NewSectionBlock(reminders, nil, nil),
				slack.NewSectionBlock(nil, reminder, nil),
				slack.NewDividerBlock(),
				slack.NewSectionBlock(exclusions, nil, nil),
				slack.NewSectionBlock(nil, requestex, nil),
			}

			// if the user is in the auth list add the exclusions command
			if helpers.Contains(b.cfg.authUsers, ctx.Event().UserID) {
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
				fmt.Println(err.Error())
			}
		},
	}
}
