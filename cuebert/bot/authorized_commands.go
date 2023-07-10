package bot

import (
	"strings"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/visual"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

func authorizationMiddleware(authorizedUserNames []string) slacker.CommandMiddlewareHandler {
	return func(next slacker.CommandHandler) slacker.CommandHandler {
		return func(ctx *slacker.CommandContext) {
			if helpers.Contains(authorizedUserNames, ctx.Event().UserID) {
				next(ctx)
			}
		}
	}
}

// app lifecycle commands
//
// initSettings allows authorized users to set the initial configuration for the bot.
func (b *Bot) initSettings() {
	definition := &slacker.CommandDefinition{
		Command:     "start cuebert",
		Description: "Set the initial configuration for the bot",
		Examples:    []string{"start cuebert"},
		Middlewares: []slacker.CommandMiddlewareHandler{authorizationMiddleware(b.cfg.authUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			b.loadPrompt(ctx.Event().UserID, Start)
		},
	}
	b.bot.AddCommand(definition)
}

// stopper allows us to stop cuebert from running via slack. requires secondary approval
func (b *Bot) stopper() {
	definition := &slacker.CommandDefinition{
		Command:     "stop cuebert",
		Description: "Stop cuebert",
		Middlewares: []slacker.CommandMiddlewareHandler{authorizationMiddleware(b.cfg.authUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			b.stopRequest(ctx.Event().UserID)
		},
		HideHelp: false,
	}
	b.bot.AddCommand(definition)
}

// updateConfig allows authorized users to update the configuration for the bot.
func (b *Bot) updateConfig() {
	definition := &slacker.CommandDefinition{
		Command:     "update config",
		Description: "Update the configuration",
		Examples:    []string{"update config"},
		Middlewares: []slacker.CommandMiddlewareHandler{authorizationMiddleware(b.cfg.authUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			b.loadPrompt(ctx.Event().UserID, Reload)
		},
	}

	b.bot.AddCommand(definition)
}

// exclusion commands
//
// addexclusion adds a device to be excluded. Only authorized users can add exclusions.
func (b *Bot) addExclusion() {
	definition := &slacker.CommandDefinition{
		Command:     "add exclusion",
		Description: "Add a device to be excluded",
		Examples:    []string{"add exclusion"},
		Middlewares: []slacker.CommandMiddlewareHandler{authorizationMiddleware(b.cfg.authUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			_, err := ctx.Response().ReplyBlocks(
				[]slack.Block{
					slack.NewSectionBlock(
						slack.NewTextBlockObject("mrkdwn", "Would you like to add a device to the exclusions list?", false, false), nil, nil),
					slack.NewActionBlock(
						ExclusionQuestion,
						slack.NewButtonBlockElement(
							YesAddExclusion,
							YesAddExclusion,
							slack.NewTextBlockObject("plain_text", Yes, false, false)).
							WithStyle(slack.StylePrimary),
						slack.NewButtonBlockElement(
							ExclusionNo,
							ExclusionNo,
							slack.NewTextBlockObject("plain_text", No, false, false)).
							WithStyle(slack.StyleDanger)),
				},
			)
			if err != nil {
				b.log.Debug().AnErr("sending exclusion prompt", err).
					Send()
			}
		},
	}

	b.bot.AddCommand(definition)
}

// requestReport returns reports about the fleet
func (b *Bot) requestReport() {
	var reportOpts = []string{"os", "manager alerted", "first message sent", "requested reminder"}

	definition := &slacker.CommandDefinition{
		Command:     "get report <opt>",
		Description: "Get reports about the fleet",
		Examples:    []string{"get report os"},
		Middlewares: []slacker.CommandMiddlewareHandler{authorizationMiddleware(b.cfg.authUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			opt := ctx.Request().Param("opt")
			var (
				vis *visual.PieChartOption
				err error
			)

			switch strings.ToLower(opt) {
			case "os":
				vis, err = b.BuildOSReport()

			case "manager alerted":
				vis, err = b.BuildSentReport(Manager)

			case "first message sent":
				vis, err = b.BuildSentReport(First)

			case "requested reminder":
				vis, err = b.BuildSentReport(ReminderRequested)

			default:
				msg := fuzzyMatchNonOpt(opt, reportOpts)
				_, err := ctx.Response().Reply(msg)

				if err != nil {
					b.log.Debug().AnErr("sending report", err).
						Send()
				}
				return
			}

			if err != nil {
				b.log.Debug().AnErr("building report", err).
					Send()
				return
			}

			err = b.sendReport(vis, ctx.Event().UserID)
			if err != nil {
				b.log.Debug().AnErr("sending report", err).
					Send()
			}
		},
	}
	b.bot.AddCommand(definition)
}

// user commands
//
// getUsersInfo returns information about a user when requested by an admin.
func (b *Bot) getUsersInfo() {
	var reportOpts = []string{"slackid", "email"}
	definition := &slacker.CommandDefinition{
		Command:     "get users info <opt> {input}",
		Description: "Get info about a user",
		Examples:    []string{"get users info slackid <id>", "get users info email <email>"},
		Handler: func(ctx *slacker.CommandContext) {
			opt := ctx.Request().Param("opt")
			which := ctx.Request().Param("input")

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
				user, err = b.tables.UserBySlackID(which)
			case "email":
				which = helpers.ExtractEmails(which)
				user, err = b.tables.UserEmail(which)
			default:
				msg := fuzzyMatchNonOpt(opt, reportOpts)
				_, err := ctx.Response().Reply(msg)

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
				ui, err := b.User("User Info", email)
				if err != nil {
					b.log.Err(err).Msg("error getting user")
				}
				attachments = append(attachments, *ui)

				bi := b.Bot("Bot Results Table", user)
				attachments = append(attachments, *bi)

				ei, err := b.Exclusions(email)
				if err != nil {
					b.log.Err(err).Msg("error getting exclusions")
				}
				attachments = append(attachments, ei...)

				di, err := b.Device(email)
				if err != nil {
					b.log.Err(err).Msg("error getting devices")
				}
				attachments = append(attachments, di...)
			}

			_, err = ctx.Response().Reply(ctx.Event().UserID, slacker.WithAttachments(attachments))
			if err != nil {
				b.log.Err(err).Msg("error responding")
			}
		},
	}

	b.bot.AddCommand(definition)
}

// updateUserInfoInteractive allows authorized users to set the initial configuration for the bot.
func (b *Bot) updateUserInfoInteractive() {
	definition := &slacker.CommandDefinition{
		Command:     "update user interactive",
		Description: "Update info about a user",
		Examples:    []string{"update user interactive"},
		Middlewares: []slacker.CommandMiddlewareHandler{authorizationMiddleware(b.cfg.authUsers)},
		Handler: func(ctx *slacker.CommandContext) {
			b.wantUpdateUser(ctx.Event().ChannelID)
		},
	}

	b.bot.AddCommand(definition)
}
