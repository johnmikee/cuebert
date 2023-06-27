package db

import (
	"time"

	"github.com/johnmikee/cuebert/db/bot"
	"github.com/rs/zerolog/log"
)

// AddACK adds the ack to the bot_results table for the given slack id
func (b *Config) AddACK(id string, t time.Time) error {
	_, err := b.br(b.db, &b.log).Update().
		FirstMessageSent(true).
		SlackID(id).
		FirstACKTime(t).
		Send()

	return err
}

// ACKACKD sets the first message ack'd flag to true for the given slack id
func (b *Config) ACKACKD(id string, t time.Time) error {
	_, err := b.br(b.db, &b.log).Update().
		FirstACK(true).
		FirstACKTime(t).
		Parse("slack_id", id).
		Send()

	return err
}

// AddManagerID adds the manager id to the bot_results table for the given slack id or email
func (b *Config) AddManagerID(slackID, userEmail, managerSlackID string) error {
	base := b.br(b.db, &b.log).Update().ManagerID(managerSlackID)

	if userEmail != "" {
		base = base.UserEmail(userEmail).Parse("user_email", userEmail)
	}

	if slackID != "" {
		base = base.SlackID(slackID).Parse("slack_id", slackID)
	}

	_, err := base.Send()

	return err
}

// BatchAddBotInfo adds all the bot info to the bot_results table
func (b *Config) BatchAddBotInfo(br bot.BR) error {
	_, err := b.br(b.db, &b.log).AddAllDevices(br)

	return err
}

// FirstMessageSent sets the first message sent flag to true
func (b *Config) FirstMessageSent(id, serial string, t time.Time) error {
	_, err := b.br(b.db, &b.log).Update().
		FirstMessageSent(true).
		FirstMessageSentAt(t).
		SlackID(id).
		Serial(serial).
		Parse("serial_number", serial).
		Send()

	return err
}

type Method string

var (
	Getter Method = "get"
	Setter Method = "set"
)

// FirstMessageWaiting sets or gets the first message waiting value
func (b *Config) FirstMessageWaiting(method Method, serial string) (bool, error) {
	switch method {
	case Getter:
		return b.getFirstMessageWaiting(serial)
	case Setter:
		return true, b.setMessageWaiting(serial)
	default:
		return false, nil
	}
}

func (b *Config) getFirstMessageWaiting(serial string) (bool, error) {
	br, err := b.br(b.db, &b.log).Query().Serial(serial).Query()
	if err != nil {
		return false, err
	}
	return br[0].FirstMessageWaiting, nil
}

func (b *Config) setMessageWaiting(serial string) error {
	_, err := b.br(b.db, &b.log).Update().
		FirstMessageWaiting(true).
		Serial(serial).
		Parse("serial_number", serial).
		Send()

	return err
}

// GetACKd returns true if the first ack flag is true
func (b *Config) GetACKd() (bot.BR, error) {
	br, err := b.br(b.db, &b.log).Query().
		All().
		Query()
	if err != nil {
		return nil, err
	}

	return br, nil
}

// GetACKTime returns the first ack time for the given serial
func (b *Config) GetACKTime(serial string) (time.Time, error) {
	br, err := b.br(b.db, &b.log).Query().Serial(serial).Query()
	if err != nil {
		return time.Time{}, err
	}

	for i := range br {
		ts := br[i].FirstACKTime
		log.Trace().
			Str("serial", serial).
			Time("first ack time", ts).
			Msg("first ack time")
		return ts, nil
	}

	return time.Time{}, nil
}

// GetBotTableInfo returns all the rows in the bot_results table
func (b *Config) GetBotTableInfo() (bot.BR, error) {
	res, err := b.br(b.db, &b.log).Query().All().Query()

	return res, err
}

// GetBotTableInfoEmail returns all the rows in the bot_results table for the given email
func (b *Config) GetBotTableInfoEmail(email string) (bot.BR, error) {
	res, err := b.br(b.db, &b.log).Query().UserEmail(email).Query()

	return res, err
}

// GetManagerNotified returns true if the manager has been notified
func (b *Config) GetManagerNotified(serial string) (bool, error) {
	br, err := b.br(b.db, &b.log).Query().Serial(serial).All().Query()
	if err != nil {
		return false, err
	}

	if len(br) == 0 {
		return false, nil
	}

	return br[0].ManagerMessageSent, nil
}

// GetFirstMessageSentAll returns all the rows in the bot_results table where the first message sent flag is true
func (b *Config) GetFirstMessageSentAll() (bot.BR, error) {
	return b.br(b.db, &b.log).Query().
		FirstMessageSent(true).
		Query()
}

// GetUsersSerialsBot returns all the rows in the bot_results table for the given slack id
func (b *Config) GetUsersSerialsBot(sid ...string) (bot.BR, error) {
	return b.br(b.db, &b.log).Query().
		SlackID(sid...).
		Query()

}

// UserByEmail returns all the rows in the bot_results table for the given email
func (b *Config) UserEmail(email ...string) (bot.BR, error) {
	return b.br(b.db, &b.log).Query().
		UserEmail(email...).
		Query()
}

// UserBySlackID returns all the rows in the bot_results table for the given slack id
func (b *Config) UserBySlackID(sid ...string) (bot.BR, error) {
	return b.br(b.db, &b.log).Query().
		SlackID(sid...).
		Query()
}

// UserByManagerSlackID returns all the rows in the bot_results table for the given manager id
func (b *Config) UserByManagerSlackID(sid ...string) (bot.BR, error) {
	return b.br(b.db, &b.log).Query().
		ManagerSlackID(sid...).
		Query()
}

// UserByFullName returns all the rows in the bot_results table for the given full name
func (b *Config) UserByFullName(name ...string) (bot.BR, error) {
	return b.br(b.db, &b.log).Query().
		FullName(name...).
		Query()
}

// UserTZOffset returns all the rows in the bot_results table for the given timezone offset
func (b *Config) UserTZOffset(offset ...int64) (bot.BR, error) {
	return b.br(b.db, &b.log).Query().
		TZS(offset...).
		Query()
}

// ManagerMessageSent returns all the rows in the bot_results table where the manager message sent flag is true
func (b *Config) ManagerMessageSent() (bot.BR, error) {
	return b.br(b.db, &b.log).Query().ManagerMessageSent(true).Query()
}

// ManagerNotifed sets the manager notified flag to true
func (b *Config) ManagerNotifed(sent bool, serial string) error {
	_, err := b.br(b.db, &b.log).Update().
		ManagerMessageSent(sent).
		ManagerMessageSentAt(time.Now().UTC()).
		Serial(serial).
		Parse("serial_number", serial).
		Send()
	return err
}

// NoManager returns all the rows in the bot_results table where the manager id is empty
func (b *Config) NoManager() (bot.BR, error) {
	return b.br(b.db, &b.log).Query().
		ManagerSlackID("").
		Query()
}

func (b *Config) RemoveBRBy() *bot.Remove {
	return b.br(b.db, &b.log).Remove()
}
