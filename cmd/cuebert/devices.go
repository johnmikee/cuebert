package main

import (
	"strings"
	"time"

	"github.com/johnmikee/cuebert/db/devices"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker"
	"github.com/slack-go/slack"
)

// deviceInfo will return information about the device if a user requests it.
func (b *Bot) deviceInfo() {
	deviceOpts := []string{"serial", "hostname", "model", "os"}
	definition := &slacker.CommandDefinition{
		Description: "Get information on your device",
		Examples:    []string{"get device serial"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			opt := request.Param("info")

			switch strings.ToLower(opt) {
			case "serial", "hostname", "model", "os":
				d, err := b.db.DevicesByUser(botCtx.Event().User)
				if err != nil {
					b.log.Debug().AnErr("getting devices", err).
						Str("user", botCtx.Event().User).
						Send()
					return
				}
				blocks := deviceRanger(opt, d)
				err = response.Reply("requested info", slacker.WithBlocks(blocks))
				if err != nil {
					b.log.Trace().
						AnErr("building blocks", err).
						Send()
				}
			default:
				msg := fuzzyMatchNonOpt(opt, deviceOpts)
				err := response.Reply(msg)

				if err != nil {
					b.log.Debug().AnErr("sending report", err).
						Send()
				}
			}
		},
	}
	b.commands = append(b.commands, Commands{
		usage: "get device {info}",
		def:   definition,
	})
}

func deviceRanger(opt string, d []devices.DeviceInfo) []slack.Block {
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

// pull info from db and compare to mdm and update where necessary
func (b *Bot) deviceDiff(time.Time) {
	b.log.Trace().Msg("checking if we need to add any devices..")
	go b.statusHandler.updateStatus(&routineUpdate{
		routine: &RoutineStatus{
			Name:    "deviceDiff",
			Start:   time.Now().Format(time.RFC3339),
			Message: "starting device diff routine",
		},
		start:  true,
		finish: false,
		err:    false,
	}, "diff")

	md, err := b.mdm.ListDevices()
	if err != nil {
		b.log.Debug().AnErr("getting devices from mdm", err).Send()
		return
	}

	ds, versCheck, err := b.tables.gatherDiffDevicesDB()

	if err != nil {
		b.log.Debug().AnErr("getting devices from db", err).Send()
		// TODO: send error back to channel
		return
	}

	b.log.Debug().Msg("checking for missing devices")
	updates := checkMissingDevices(ds, b.cfg.flags.requiredVers, md)

	b.log.Debug().Msg("checking for devices needing to be removed")
	remove := checkStaleDevices(b.cfg.flags.requiredVers, versCheck, md)

	b.log.Trace().Interface("devices", updates).Msg("adding devices")

	err = b.db.AddAll(updates)
	if err != nil {
		b.log.Debug().AnErr("adding devices", err).Send()
	}

	b.log.Trace().Strs("devices", remove).Msg("removing devices")
	_, err = b.db.RemoveDeviceBy().Serial(remove...).Execute()
	if err != nil {
		b.log.Debug().AnErr("removing devices", err).Send()
	}

	go b.statusHandler.updateStatus(&routineUpdate{
		routine: &RoutineStatus{
			Name:          "deviceDiff",
			Finish:        time.Now().Format(time.RFC3339),
			Message:       "finished device diff routine",
			FinishNoError: true,
		},
		start:  false,
		finish: true,
		err:    false,
	}, "diff")
}

// checkMissingDevices checks for devices that are missing by polling
// the MDM and comparing the results to the DB. If the device is missing
// and the OS version is lower than the required version, it will be added.
func checkMissingDevices(ds []string, reqVers string, mdmDevices mdm.DeviceResults) []devices.DeviceInfo {
	updates := []devices.DeviceInfo{}

	for i := range mdmDevices {
		// check missing
		if !helpers.Contains(ds, mdmDevices[i].SerialNumber) {
			ok, _ := helpers.CompareOSVer(mdmDevices[i].OSVersion, reqVers)
			if !ok {
				// check that the platform is macOS
				// V2: Add multiple OS support
				if mdmDevices[i].Platform == "Mac" {
					di := devices.DeviceInfo{
						DeviceID:     mdmDevices[i].DeviceID,
						DeviceName:   mdmDevices[i].DeviceName,
						Model:        mdmDevices[i].Model,
						SerialNumber: mdmDevices[i].SerialNumber,
						Platform:     mdmDevices[i].Platform,
						OSVersion:    mdmDevices[i].OSVersion,
						LastCheckIn:  mdmDevices[i].LastCheckIn,
						User:         mdmDevices[i].User.Email,
						UserMDMID:    mdmDevices[i].User.ID,
					}

					updates = append(updates, di)
				}
			}
		}
	}

	return updates
}

type check struct {
	serial string
	os     string
}

// checkStaleDevices checks for devices that no longer need to be in the DB.
// If the OS on the device is greater than the required version, it will be removed.
func checkStaleDevices(reqVers string, versCheck []check, md mdm.DeviceResults) []string {
	remove := []string{}

	for i := range md {
		for _, x := range versCheck {
			if md[i].SerialNumber == x.serial {
				ok, _ := helpers.CompareOSVer(x.os, md[i].OSVersion)
				// if the MDM version is greater than the DB version
				// double check that its not above the required
				if ok {
					doubleCheck, _ := helpers.CompareOSVer(reqVers, x.os)
					if !doubleCheck {
						remove = append(remove, x.serial)
					}
				} else {
					ok, _ := helpers.CompareOSVer(reqVers, x.os)
					if ok {
						remove = append(remove, x.serial)
					}
				}
			}
		}
	}

	return remove
}
