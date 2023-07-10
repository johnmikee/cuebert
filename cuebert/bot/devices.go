package bot

import (
	"strings"

	"github.com/johnmikee/cuebert/db/devices"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// deviceInfo will return information about the device if a user requests it.
func (b *Bot) deviceInfo() {
	deviceOpts := []string{"serial", "hostname", "model", "os"}
	definition := &slacker.CommandDefinition{
		Command:     "get device {info}",
		Description: "Get information on your device",
		Examples:    []string{"get device serial"},
		Handler: func(ctx *slacker.CommandContext) {
			opt := ctx.Request().Param("info")

			switch strings.ToLower(opt) {
			case "serial", "hostname", "model", "os":
				d, err := b.tables.DevicesByUser(ctx.Event().UserID)
				if err != nil {
					b.log.Debug().AnErr("getting devices", err).
						Str("user", ctx.Event().UserID).
						Send()
					return
				}
				blocks := deviceRanger(opt, d)
				_, err = ctx.Response().ReplyBlocks(blocks)
				if err != nil {
					b.log.Trace().
						AnErr("building blocks", err).
						Send()
				}
			default:
				msg := fuzzyMatchNonOpt(opt, deviceOpts)
				_, err := ctx.Response().Reply(msg)

				if err != nil {
					b.log.Debug().AnErr("sending report", err).
						Send()
				}
			}
		},
	}
	b.bot.AddCommand(definition)
}

func deviceRanger(opt string, d devices.DI) []slack.Block {
	resp := []string{}

	for i := range d {
		switch strings.ToLower(opt) {
		case "serial":
			resp = append(resp, d[i].SerialNumber)
		case "hostname":
			resp = append(resp, d[i].DeviceName)
		case "model":
			resp = append(resp, d[i].Model)
		case "os":
			resp = append(resp, d[i].OSVersion)
		}
	}

	return deviceResp(resp)
}

func deviceResp(opts []string) []slack.Block {
	query := "Here are the results I found for you:\n"
	header := slack.NewTextBlockObject(slack.MarkdownType, query, false, false)

	found := []*slack.TextBlockObject{}

	found = append(found,
		slack.NewTextBlockObject(slack.MarkdownType, strings.Join(opts, "\n"), false, false))

	return []slack.Block{
		slack.NewSectionBlock(header, nil, nil),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(nil, found, nil),
		slack.NewDividerBlock(),
	}

}
