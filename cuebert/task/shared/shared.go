package shared

import (
	"time"

	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/bot/lifecycle"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task/info"

	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

type Shared struct {
	bot       *bot.Bot
	config    *config.Config
	flags     *config.Flags
	db        db.Client
	log       logger.Logger
	status    *handlers.StatusHandler
	lifecycle *lifecycle.Client
	caller    Caller
	impl      ImplChecker
}

type Config struct {
	Config      *config.Config
	Bot         *bot.Bot
	LifeCycle   *lifecycle.Client
	DB          db.Client
	Flags       *config.Flags
	Log         logger.Logger
	Status      *handlers.StatusHandler
	Caller      Caller
	ImplChecker ImplChecker
}

type Caller interface {
	CheckNotSent(serviceID, slackID string) (bool, error)
	FirstMessageSent(serviceID, slackID string) error
	FirstMessageDelivered(serviceID, slackID string) error
	FinalMessageSent(serviceID, slackID string) error
	IsExcluded(string) (bool, bool) // returns if exluded and if approved
	ReminderMessage(*ReminderPayload) error
	MessageAcknowledged(userID string, when time.Time, which string) error
	ExclusionRequestDecision(ctx *slacker.InteractionContext)
	ExclusionYes(ctx *slacker.InteractionContext)
	ExclusionAddBlocks() []*slack.InputBlock
	ExclusionSubmit(ctx *slacker.InteractionContext)
	GetUserTZ(slackID string) (int64, error)
	UserExclusionSubmit(ctx *slacker.InteractionContext)
	UpdateReminderTime(dateVal, timeVal, slackID string) error
	UserUpdate(ctx *slacker.InteractionContext)
	UserUpdateSubmit(ctx *slacker.InteractionContext)
	UserCompliant(string) error
}

type ImplChecker interface {
	CheckComplianceStatus(user string) (bool, error)
	CheckComplianceAll() ([]info.PersistUsers, error)
}

// New will return a new User struct.
func New(u *Config) *Shared {
	return &Shared{
		config:    u.Config,
		bot:       u.Bot,
		flags:     u.Flags,
		db:        u.DB,
		log:       logger.ChildLogger("shared", &u.Log),
		lifecycle: u.LifeCycle,
		status:    u.Status,
		caller:    u.Caller,
		impl:      u.ImplChecker,
	}
}

func (s *Shared) UpdateStatus(up *handlers.RoutineUpdate, which string) {
	s.status.UpdateStatus(up, which)
}

type Options func(*Shared) *Shared

func (s *Shared) Update(opt ...Options) *Shared {
	for _, o := range opt {
		s = o(s)
	}
	return s
}

func WithConfig(c *config.Config) Options {
	return func(s *Shared) *Shared {
		s.config = c
		return s
	}
}

func WithBot(b *bot.Bot) Options {
	return func(s *Shared) *Shared {
		s.bot = b
		return s
	}
}

func WithFlags(f *config.Flags) Options {
	return func(s *Shared) *Shared {
		s.flags = f
		return s
	}
}

func WithDB(d db.Client) Options {
	return func(s *Shared) *Shared {
		s.db = d
		return s
	}
}

func WithLog(l logger.Logger) Options {
	return func(s *Shared) *Shared {
		s.log = l
		return s
	}
}

func WithStatus(h *handlers.StatusHandler) Options {
	return func(s *Shared) *Shared {
		s.status = h
		return s
	}
}

func WithLifeCycle(l *lifecycle.Client) Options {
	return func(s *Shared) *Shared {
		s.lifecycle = l
		return s
	}
}

func WithCaller(c Caller) Options {
	return func(s *Shared) *Shared {
		s.caller = c
		return s
	}
}

func WithImplChecker(i ImplChecker) Options {
	return func(s *Shared) *Shared {
		s.impl = i
		return s
	}
}

func (s *Shared) ImplChecker(i ImplChecker) *Shared {
	s.impl = i
	return s
}
