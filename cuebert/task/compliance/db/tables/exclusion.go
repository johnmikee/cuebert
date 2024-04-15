package tables

import (
	"errors"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/johnmikee/cuebert/db/shared/exclusions"
)

// Request allows a user to request an Exclusion
func (c *Config) Request(slackID, reason string, until time.Time) error {
	user, err := c.Tracking().By().Opts(tracking.UserSlackID, slackID).Query()
	if err != nil {
		return err
	}

	if user.Empty() {
		return errors.New("no record found with that slack id")
	}

	_, err = c.Exclusion().Add().WithOptions(
		exclusions.Info{
			Approved:  false,
			UserEmail: user[0].UserEmail,
			Reason:    reason,
			Until:     until,
		},
	).Execute()

	return err
}
