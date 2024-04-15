package common

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
	"github.com/johnmikee/cuebert/db/shared/users"
)

// GetUserTZ returns the timezone offset for the given slack id
func (c *Common) GetUserTZ(slackID string) (int64, error) {
	ui, err := c.tables.User().By().Opts(users.UserSlackID, slackID).Query()
	if err != nil {
		return 0, err
	}

	if ui.Empty() {
		return 0, fmt.Errorf("no user found for slack id %s", slackID)
	}
	return ui[0].TZOffset, nil
}

// MessageAcknowledged updates the tracking table with the given information
func (c *Common) MessageAcknowledged(userID string, when time.Time, which string) error {
	info := tracking.Info{
		UserSlackID: userID,
	}
	switch which {
	case "first":
		info.FirstACK = true
		info.FirstACKTime = when
	case "final":
		info.FinalMessageSent = true
		info.FinalMessageSentAt = when
	case "reminder":
		c.log.Info().Str("which", which).Msg("reminder ack")
	default:
		c.log.Info().Str("which", which).Msg("unknown ack type")
		return nil
	}
	_, err := c.tables.Tracking().Update().WithOptions(info).Parse(tracking.UserSlackID, userID).Send()

	return err
}
