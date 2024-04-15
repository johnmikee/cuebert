package tables

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/rs/zerolog/log"
)

type Method string

var (
	Getter Method = "get"
	Setter Method = "set"
)

// FirstMessageWaiting sets or gets the first message waiting value
func (c *Config) FirstMessageWaiting(method Method, serial string) (bool, error) {
	switch method {
	case Getter:
		return c.getFirstMessageWaiting(serial)
	case Setter:
		return true, c.setMessageWaiting(serial)
	default:
		return false, nil
	}
}

func (c *Config) getFirstMessageWaiting(serial string) (bool, error) {
	br, err := c.Tracking().By().Opts(tracking.SerialNumber, serial).Query()
	if err != nil {
		return false, err
	}
	return br[0].FirstMessageWaiting, nil
}

func (c *Config) setMessageWaiting(serial string) error {
	_, err := c.Tracking().Update().WithOptions(tracking.Info{
		FirstMessageWaiting: true,
		SerialNumber:        serial,
	}).Parse(tracking.SerialNumber, serial).Send()

	return err
}

// GetACKTime returns the first ack time for the given serial
func (c *Config) GetACKTime(serial string) (time.Time, error) {
	br, err := c.Tracking().By().Opts(tracking.SerialNumber, serial).Query()
	if err != nil {
		return time.Time{}, err
	}

	if br.Empty() {
		return time.Time{}, fmt.Errorf("no results found for serial %s", serial)
	}

	ts := br[0].FirstACKTime
	log.Trace().
		Str("serial", serial).
		Time("first ack time", ts).
		Msg("first ack time")

	return ts, nil
}

// GetBotTableInfo returns all the rows in the bot_results table
func (c *Config) GetTableInfo() (tracking.TI, error) {
	res, err := c.Tracking().By().All().Query()

	return res, err
}

// GetBotTableInfoEmail returns all the rows in the bot_results table for the given email
func (c *Config) GetBotTableInfoEmail(email string) (tracking.TI, error) {
	res, err := c.Tracking().By().Opts(tracking.UserEmail, email).Query()

	return res, err
}

// GetReminderInterval returns the reminder interval for the given serial
func (c *Config) GetReminderInterval(serial string) (int, error) {
	br, err := c.Tracking().By().Opts(tracking.SerialNumber, serial).Query()
	if err != nil {
		return 0, err
	}

	if br.Empty() {
		return 0, fmt.Errorf("no results found for serial %s", serial)
	}

	return br[0].ReminderInterval, nil
}

// GetReminderWaiting returns true if the reminder waiting flag is true for the given serial
func (c *Config) GetReminderWaiting(serial string) (bool, error) {
	br, err := c.Tracking().By().Opts(tracking.SerialNumber, serial).Query()
	if err != nil {
		return false, err
	}

	if br.Empty() {
		return false, fmt.Errorf("no results found for serial %s", serial)
	}

	return br[0].ReminderWaiting, nil
}
