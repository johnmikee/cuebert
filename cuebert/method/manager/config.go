package manager

import (
	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/cuebert/tables"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
)

type Config struct {
	SlackAlertChannel string
	Log               logger.Logger
	DB                *tables.Config
	Bot               *bot.Bot
	IDP               idp.Provider
	MDM               mdm.Provider
	Handler           *handlers.StatusHandler
}

type Cfg struct {
	slackAlertChannel string   // the channel to send alerts to
	cutoffTime        string   // cutoffTime will be the time access is revoked
	deadline          string   // the day the update is required
	requiredVers      string   // ex: 13.1
	testing           bool     // run in testing mode
	testingUsers      []string // array of users to test with
	pollInterval      int      // how often to poll the idp for users
}

type Option func(*Cfg)

func WithOptions(opts ...Option) *Cfg {
	cfg := &Cfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func WithSlackAlertChannel(channel string) Option {
	return func(cfg *Cfg) {
		cfg.slackAlertChannel = channel
	}
}

func WithCutoffTime(time string) Option {
	return func(cfg *Cfg) {
		cfg.cutoffTime = time
	}
}

func WithDeadline(time string) Option {
	return func(cfg *Cfg) {
		cfg.deadline = time
	}
}

func WithRequiredVers(vers string) Option {
	return func(cfg *Cfg) {
		cfg.requiredVers = vers
	}
}

func WithTesting(testing bool) Option {
	return func(cfg *Cfg) {
		cfg.testing = testing
	}
}

func WithTestingUsers(users []string) Option {
	return func(cfg *Cfg) {
		cfg.testingUsers = users
	}
}

func WithPollInterval(interval int) Option {
	return func(cfg *Cfg) {
		cfg.pollInterval = interval
	}
}
