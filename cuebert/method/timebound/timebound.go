package timebound

import (
	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/cuebert/method"
	"github.com/johnmikee/cuebert/cuebert/tables"
	br "github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/slack-go/slack"
)

type TimeBound struct {
	log           logger.Logger
	tables        *tables.Config
	bot           *bot.Bot
	intervals     Intervals
	cfg           *Cfg
	sc            *slack.Client
	statusHandler *handlers.StatusHandler
}

type Intervals struct {
	base   br.BR
	thirty br.BR
	hour   br.BR
	two    br.BR
	four   br.BR
}

const (
	ReminderMessage = "reminder_message"
)

// TableAssociations implements method.Actions.
func (t *TimeBound) TableAssociations([]string) {
	t.log.Trace().Msg("nothing to do")
}

// Setup implements method.Actions.
func (t *TimeBound) Setup(method method.Config) {
	t.log = method.Log
	t.tables = method.Tables
	t.bot = method.Bot
	t.sc = method.SlackClient
	t.statusHandler = method.StatusHandler
	t.cfg = WithOptions(
		WithCutoffTime(method.CutoffTime),
		WithDeadline(method.Deadline),
		WithRequiredVers(method.RequiredVers),
		WithSlackAlertChannel(method.SlackAlertChannel),
		WithTesting(method.Testing),
		WithTestingUsers(method.TestingUsers),
		WithPollInterval(method.PollInterval),
	)
}
