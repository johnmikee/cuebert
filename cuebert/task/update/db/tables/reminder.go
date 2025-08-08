package tables

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
)

// PullReminderInfo returns all the rows in the bot_results table.
func (b *Config) PullReminderInfo() (tracking.TI, error) {
	return b.Tracking().By().All().Query()
}

// ReminderSent sets the reminder sent flag to true.
func (b *Config) ReminderSent(sent bool, serial string) error {
	_, err := b.Tracking().Update().
		WithOptions(tracking.Info{ReminderSent: sent, SerialNumber: serial}).
		Parse(tracking.SerialNumber, serial).Send()

	return err
}

// ReminderSentCheck returns true if the reminder has been sent.
func (b *Config) ReminderSentCheck(s string) (bool, error) {
	res, err := b.Tracking().By().Opts(tracking.UserSlackID, s).Query()
	if err != nil {
		return false, err
	}

	if res.Empty() {
		return false, fmt.Errorf("no results found for slack id %s", s)
	}

	return res[0].ReminderSent, nil
}

// SetDate sets the reminder date for the given slack id.
func (b *Config) SetDate(id, date string) error {
	_, err := b.Tracking().Update().
		WithOptions(tracking.Info{UserReminderDate: date}).
		Parse(tracking.UserSlackID, id).Send()

	return err
}

// SetReminder sets the reminder time for the given slack id.
func (b *Config) SetReminder(id string, t time.Time, ds, ts string) error {
	_, err := b.Tracking().Update().
		WithOptions(
			tracking.Info{
				UserReminderAt:   t,
				UserReminderDate: ds,
				UserReminderTime: ts,
			},
		).
		Parse(tracking.UserSlackID, id).Send()

	return err
}

// SetTime sets the reminder time for the given slack id.
func (b *Config) SetTime(id, t string) error {
	_, err := b.tracking(b.connection).Update().
		WithOptions(
			tracking.Info{UserReminderTime: t},
		).
		Parse(tracking.UserSlackID, id).Send()

	return err
}
