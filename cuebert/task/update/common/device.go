package common

import (
	"strconv"
	"strings"
	"time"

	"github.com/johnmikee/cuebert/bot/ops"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/devices"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tables"
	"github.com/johnmikee/cuebert/db/shared/updates"
	"github.com/johnmikee/cuebert/db/shared/users"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

type OS struct {
	OS    string
	Count int
}

func (c *Common) OSCount() OSCount {
	all, err := c.tables.Device().By().All().Query()
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("could not query devices")
		return nil
	}

	os := make(map[string]int)
	for _, device := range all {
		os[device.OSVersion]++
	}

	var osCount []OS
	for k, v := range os {
		osCount = append(osCount, OS{
			OS:    k,
			Count: v,
		})
	}

	return osCount
}

type OSCount []OS

type MajorBreakdown struct {
	MajorVers string
	OS        []OS
}

// ByMajorVers will group the OS by major version
func (o OSCount) ByMajorVers() []MajorBreakdown {
	major := make(map[string][]OS)
	for _, os := range o {
		// Split the version string by dot and take the first part as the major version
		parts := strings.Split(os.OS, ".")
		majorVersion := parts[0]

		major[majorVersion] = append(major[majorVersion], os)
	}

	var majorBreakdown []MajorBreakdown
	for k, v := range major {
		majorBreakdown = append(majorBreakdown, MajorBreakdown{
			MajorVers: k,
			OS:        v,
		})
	}

	// now sort the major breakdown by major version from lowest to highest
	for i := 0; i < len(majorBreakdown); i++ {
		for j := i + 1; j < len(majorBreakdown); j++ {
			if majorBreakdown[i].MajorVers > majorBreakdown[j].MajorVers {
				majorBreakdown[i], majorBreakdown[j] = majorBreakdown[j], majorBreakdown[i]
			}
		}
	}
	return majorBreakdown
}

// AddUpdates adds the OS count to the updates table
func (o OSCount) AddUpdates(tables *tables.Config) error {
	_, week := time.Now().ISOWeek()
	day := int(time.Now().Weekday())

	// use a batch add to add all the updates at once
	ui := []updates.Info{}
	for _, m := range o {
		u := updates.Info{
			Week:     week,
			ItemName: m.OS,
		}
		switch day {
		case 1:
			u.Monday = m.Count
		case 2:
			u.Tuesday = m.Count
		case 3:
			u.Wednesday = m.Count
		case 4:
			u.Thursday = m.Count
		case 5:
			u.Friday = m.Count
		}

		ui = append(ui, u)
	}

	_, err := tables.Updates().Add().BatchAdd(ui)
	if err != nil {
		return err
	}

	return nil
}

// this makes an assumption the user field is set as email which may not be true in all cases.
func (c *Common) Device(email string) ([]slack.Attachment, error) {
	devices, err := c.tables.Device().By().Opts(devices.User, email).Query()
	if err != nil {
		return nil, err
	}

	if devices.Empty() {
		return nil, nil
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

// DeviceInfo will return information about the device if a user requests it.
func (c *Common) DeviceInfo() {
	deviceOpts := []string{"serial", "hostname", "model", "os"}
	definition := &slacker.CommandDefinition{
		Command:     "get device {info}",
		Description: "Get information on your device",
		Examples:    []string{"get device serial"},
		Handler: func(ctx *slacker.CommandContext) {
			opt := ctx.Request().Param("info")

			switch strings.ToLower(opt) {
			case "serial", "hostname", "model", "os":
				c.log.Info().Str("opt", opt).Send()
				user, err := c.tables.User().By().Opts(users.UserSlackID, ctx.Event().UserID).Query()
				if err != nil {
					c.log.Debug().AnErr("getting user", err).
						Str("user", ctx.Event().UserID).
						Send()
					return
				}
				if user.Empty() {
					c.log.Debug().Msg("user not found")
					return
				}

				uid := user.GetServiceID()
				d, err := c.tables.Device().By().Opts(devices.UserMDMID, strconv.Itoa(uid)).Query()
				if err != nil {
					c.log.Debug().AnErr("getting devices", err).
						Str("user", ctx.Event().UserID).
						Send()
					return
				}
				blocks := deviceRanger(opt, d)
				c.log.Info().Interface("blocks", blocks).Send()
				_, err = ctx.Response().ReplyBlocks(blocks)
				if err != nil {
					c.log.Trace().
						AnErr("building blocks", err).
						Send()
				}
			default:
				msg := ops.FuzzyMatchNonOpt(opt, deviceOpts)
				_, err := ctx.Response().Reply(msg)
				if err != nil {
					c.log.Debug().AnErr("sending report", err).
						Send()
				}
			}
		},
	}
	if c.bot == nil {
		c.log.Debug().Msg("bot is nil")
		return
	}
	c.bot.Bot().AddCommand(definition)
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
