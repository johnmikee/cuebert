package db

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/db/bot"
)

// PullReminderInfo returns all the rows in the bot_results table
func (b *Config) PullReminderInfo() ([]bot.BotResInfo, error) {
	return b.br(b.db, &b.log).Query().All().Query()
}

// ReminderSent sets the reminder sent flag to true
func (b *Config) ReminderSent(sent bool, serial string) error {
	_, err := b.br(b.db, &b.log).Update().
		DelaySent(sent).
		Parse("serial_number", serial).
		Send()

	return err
}

// ReminderSentCheck returns true if the reminder has been sent
func (b *Config) ReminderSentCheck(s string) (bool, error) {
	res, err := b.br(b.db, &b.log).Query().SlackID(s).Query()

	if err != nil {
		return false, err
	}

	if len(res) == 0 {
		return false, fmt.Errorf("no bot found with slack id %s", s)
	}

	return res[0].DelaySent, nil
}

// SetDate sets the reminder date for the given slack id
func (b *Config) SetDate(id, date string) error {
	_, err := b.br(b.db, &b.log).Update().
		DelayDate(date).
		Parse("slack_id", id).
		Send()

	return err
}

// SetReminder sets the reminder time for the given slack id
func (b *Config) SetReminder(id string, t time.Time, ds, ts string) error {
	_, err := b.br(b.db, &b.log).Update().
		DelayAt(t).
		DelayDate(ds).
		DelayTime(ts).
		Parse("slack_id", id).
		Send()

	return err
}

// SetTime sets the reminder time for the given slack id
func (b *Config) SetTime(id, t string) error {
	_, err := b.br(b.db, &b.log).Update().
		DelayTime(t).
		Parse("slack_id", id).
		Send()

	return err
}
