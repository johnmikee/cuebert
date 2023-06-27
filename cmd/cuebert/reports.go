package main

import (
	"strings"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/johnmikee/cuebert/pkg/visual"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
)

// requestReport returns reports about the fleet
func (b *Bot) requestReport() *slacker.CommandDefinition {
	var reportOpts = []string{"os", "manager alerted", "first message sent", "requested reminder"}

	definition := &slacker.CommandDefinition{
		Description: "Get reports about the fleet",
		Examples:    []string{"get report"},
		AuthorizationFunc: func(botCtx slacker.BotContext, request slacker.Request) bool {
			return helpers.Contains(b.cfg.authUsers, botCtx.Event().User)
		},
		HideHelp: false,
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			opt := request.Param("opt")
			var (
				vis *visual.PieChartOption
				err error
			)

			switch strings.ToLower(opt) {
			case "os":
				vis, err = b.tables.buildOSReport()

			case "manager alerted":
				vis, err = b.tables.buildSentReport(ManagerMessageSent)

			case "first message sent":
				vis, err = b.tables.buildSentReport(FirstMessageSent)

			case "requested reminder":
				vis, err = b.tables.buildSentReport(ReminderRequested)

			default:
				msg := fuzzyMatchNonOpt(opt, reportOpts)
				err := response.Reply(msg)

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

			err = b.sendReport(vis, botCtx.Event().User)
			if err != nil {
				b.log.Debug().AnErr("sending report", err).
					Send()
			}
		},
	}
	return definition
}

func countSentStatus(br []bot.BotResInfo, which Report) (sent, notSent int) {
	for i := range br {
		switch which {
		case FirstMessageSent:
			if br[i].FirstMessageSent {
				sent++
			} else {
				notSent++
			}
		case ManagerMessageSent:
			if br[i].ManagerMessageSent {
				sent++
			} else {
				notSent++
			}
		case ReminderRequested:
			if !br[i].DelayAt.IsZero() {
				sent++
			} else {
				notSent++
			}
		}
	}
	return sent, notSent
}

func generateReport(fileName, title string) slack.FileUploadParameters {
	return slack.FileUploadParameters{
		Title:    title,
		Filetype: "image/png",
		File:     fileName,
	}
}

func (b *Bot) sendReport(o *visual.PieChartOption, channel ...string) error {
	graph, err := visual.PieChart(o)

	if err != nil {
		return err
	}

	report := generateReport(graph, o.Text)
	report.Channels = channel

	_, err = b.bot.Client().UploadFile(report)

	return err
}

// SendDailyReport sends a daily report to the configured channel
// with a pie chart of the OS breakdown of all devices in the database.
func (b *Bot) SendDailyAdminReport() error {
	r, err := b.tables.buildOSReport()
	if err != nil {
		b.log.Debug().AnErr("building report", err).
			Send()
		return err
	}

	err = b.sendReport(r, b.cfg.SlackAlertChannel)

	if err != nil {
		b.log.Debug().AnErr("sending report", err).
			Send()
	}

	return err
}
