package method

import (
	"time"

	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/cuebert/tables"
	dbot "github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

type Option string

const (
	Manager   Option = "manager"
	TimeBound Option = "timebound"
)

type Actions interface {
	Routines
	Messaging
	Tasks
	Setup
}

type Routines interface {
	Check(time.Time)
	Poll(time.Time)
	DeviceDiff([]string)
}

type Tasks interface {
	PostCheck([]string)
	TableAssociations([]string)
	UserUpdate(ctx *slacker.InteractionContext)
	UserUpdateModalSubmit(base *dbot.Info, values map[string]map[string]slack.BlockAction)
	Deadline()
}

type Messaging interface {
	FirstMessage() string
	ReminderMessage(rp *bot.ReminderPayload) error
}

type Setup interface {
	Setup(Config)
}

type Config struct {
	Log               logger.Logger
	Tables            *tables.Config
	Bot               *bot.Bot
	StatusHandler     *handlers.StatusHandler
	SlackClient       *slack.Client
	IDP               idp.Provider
	MDM               mdm.Provider
	SlackAlertChannel string
	CutoffTime        string
	Deadline          string
	RequiredVers      string
	Testing           bool
	TestingUsers      []string
	PollInterval      int
}
