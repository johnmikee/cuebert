package main

import (
	"strings"
	"time"

	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/cuebert/tables"
	"github.com/johnmikee/cuebert/db/devices"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/helpers"
)

var checkPlatforms = []string{"mac", "macos"}

// pull info from db and compare to mdm and update where necessary
func (c *Cuebert) deviceDiff(time.Time) {
	c.log.Trace().Msg("checking if we need to add any devices..")
	go c.statusHandler.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:    "deviceDiff",
				Start:   time.Now().Format(time.RFC3339),
				Message: "starting device diff routine",
			},
			Start:  true,
			Finish: false,
			Err:    false,
		}, "diff")

	md, err := c.mdm.ListDevices()
	if err != nil {
		c.log.Debug().AnErr("getting devices from mdm", err).Send()
		return
	}

	ds, versCheck, err := c.tables.GatherDiffDevicesDB()

	if err != nil {
		c.log.Debug().AnErr("getting devices from db", err).Send()
		return
	}

	c.log.Debug().Msg("checking for missing devices")
	updates := checkMissingDevices(ds, c.flags.requiredVers, md)

	c.log.Debug().Msg("checking for devices needing to be removed")
	remove := tables.CheckStaleDevices(c.flags.requiredVers, versCheck, md)

	c.log.Trace().Interface("devices", updates).Msg("adding devices")

	err = c.tables.AddAll(updates)
	if err != nil {
		c.log.Debug().AnErr("adding devices", err).Send()
	}

	c.log.Trace().Strs("devices", remove).Msg("removing devices")
	c.method.DeviceDiff(remove)
	_, err = c.tables.RemoveDeviceBy().Serial(remove...).Execute()
	if err != nil {
		c.log.Debug().AnErr("removing devices", err).Send()
	}

	go c.statusHandler.UpdateStatus(
		&handlers.RoutineUpdate{
			Routine: &handlers.RoutineStatus{
				Name:          "deviceDiff",
				Finish:        time.Now().Format(time.RFC3339),
				Message:       "finished device diff routine",
				FinishNoError: true,
			},
			Start:  false,
			Finish: true,
			Err:    false,
		},
		"diff",
	)
}

// checkMissingDevices checks for devices that are missing by polling
// the MDM and comparing the results to the DB. If the device is missing
// and the OS version is lower than the required version, it will be added.
func checkMissingDevices(ds []string, reqVers string, mdmDevices mdm.DeviceResults) devices.DI {
	updates := devices.DI{}

	for i := range mdmDevices {
		// check missing
		if !helpers.Contains(ds, mdmDevices[i].SerialNumber) {
			ok, _ := helpers.CompareOSVer(mdmDevices[i].OSVersion, reqVers)
			if !ok {
				// check that the platform is macOS
				// V2: Add multiple OS support
				if helpers.Contains(checkPlatforms, strings.ToLower(mdmDevices[i].Platform)) {
					di := devices.Info{
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
