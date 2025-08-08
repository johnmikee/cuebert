package tables

import (
	"errors"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/update/db/devices"
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/db/shared/exclusions"
)

// AddExclusion adds an Exclusion to the database.
func (c *Config) Add(serial, reason string, until time.Time) error {
	di, err := c.Device().By().Opts(devices.SerialNumber, serial).Query()
	if err != nil {
		return err
	}

	if di.Empty() {
		return errors.New("no device found with that serial number")
	}

	email := di[0].User

	_, err = c.Exclusion().Add().WithOptions(exclusions.Info{
		Approved:  true,
		UserEmail: email,
		Reason:    reason,
		// SerialNumber: serial,
		Until: until,
	}).Execute()

	return err
}

// ApproveExclusion approves an Exclusion
// func (c *Config) Approve(reason, serial string, until time.Time) error {
// 	_, err := e.exclusions.Update().
// 		Approved(true).
// 		Reason(reason).
// 		SerialNumber(serial).
// 		Until(until).
// 		Parse("serial_number", serial).
// 		Send()

// 	return err
// }

func (c *Config) Serials(slackID string) []string {
	resp, err := c.Tracking().By().Opts(tracking.UserSlackID, slackID).Query()
	if err != nil {
		c.log.Error().Err(err).Msg("could not get serials")
		return nil
	}

	if resp.Empty() {
		c.log.Info().Msg("no serials found")
		return nil
	}

	serials := []string{}
	// for i := range resp {
	// 	serials = append(serials, resp[i].SerialNumber)
	// }

	return serials
}

// RequestExclusion allows a user to request an Exclusion.
func (c *Config) Request(slackID, reason string, serials []string, until time.Time) error {
	user, err := c.Tracking().By().Opts(tracking.UserSlackID, slackID).Query()
	if err != nil {
		return err
	}

	if user.Empty() {
		return errors.New("no record found with that slack id")
	}

	for i := range user {
		// if !helpers.Contains(serials, user[i].SerialNumber) {
		// 	continue
		// }

		_, err = c.Exclusion().Add().WithOptions(exclusions.Info{
			Approved:  false,
			UserEmail: user[i].UserEmail,
			Reason:    reason,
			// SerialNumber: user[i].SerialNumber,
			Until: until,
		}).Execute()
	}
	return err
}
