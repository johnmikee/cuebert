package task

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/bot/lifecycle"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task/implementation"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/db"
	dbc "github.com/johnmikee/cuebert/db/config"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/shomali11/slacker/v2"

	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/slack-go/slack"
)

// Task stores both the option and the implementation of the task.
type Task struct {
	Option         Option
	Implementation implementation.Item
}

// Option is the type of task to be run.
type Option string

const (
	Update     Option = "update"     // os updates
	Compliance Option = "compliance" // various compliance tasks
)

// Config is the configuration for the task.
type Config struct {
	DBConfig      *dbc.Config
	LifeCycle     *lifecycle.Client
	Log           logger.Logger
	DB            db.Client
	Pool          *pgxpool.Conn
	Bot           *bot.Bot
	StatusHandler *handlers.StatusHandler
	SlackClient   *slack.Client
	Shared        *shared.Shared
	IDP           idp.Provider
	MDM           mdm.Provider
	Flags         *config.Flags
	Config        *config.Config
}

// Rabbit is the interface for the task.
type Rabbit interface {
	Bot
	Jobs
	Lifecycle
	Messaging
	Routines
	Setup
	shared.Caller
}

// Bot handles all the actions the bot must take.
type Bot interface {
	ExclusionRequestDecision(ctx *slacker.InteractionContext)
	ExclusionYes(ctx *slacker.InteractionContext)
	ExclusionAddBlocks() []*slack.InputBlock
	ExclusionSubmit(ctx *slacker.InteractionContext)
	GetUserTZ(slackID string) (int64, error)
	UserExclusionSubmit(ctx *slacker.InteractionContext)
	UpdateReminderTime(dateVal, timeVal, slackID string) error
	UserUpdate(ctx *slacker.InteractionContext)
	UserUpdateSubmit(ctx *slacker.InteractionContext)
	InteractionHandler(ctx *slacker.InteractionContext)
	InteractiveHelper(ctx *slacker.InteractionContext)
	LoadInput(ctx *slacker.InteractionContext, loadType string)
	LoadProgram(triggerID string, loadType string)
	ModalSubmit(ctx *slacker.InteractionContext)
	RegisterHelp()
	RegisterCommands()
	RegisterAdminCommands()
	RegisterMiddlewares()
	RegisterHandlers()
}

// Jobs contain utility functions for the task.
type Jobs interface {
	LogFlags()
	SetupReportHandler() *handlers.Handler
}

// Job is used by the job scheduler to run the various routine tasks.
type Job struct {
	Job        func(time.Time) // Job is the function to be run.
	Name       string          // Name is used to identify the job.
	When       time.Duration   // When is used to determine how often the job should be run.
	Continuous bool            // Continuous is used to determine if the job should be run beyond the task deadline.
}

// Lifecycle manages if cuebert is active or not.
type Lifecycle interface {
	Start()
	Stop()
	Update()
}

// Messaging stores the messages to be sent to the user by the bot for each task.
type Messaging interface {
	FirstMessage() string
	FinalMessage() string
	FirstMessageHeader() string
}

// Routines are the functions that execute on a schedule when cuebert is active with the exception of Populate and CheckRemove.
// these two run as maintenance tasks to ensure that the database is up to date and that users are not left in Okta jail.
type Routines interface {
	Get() []Job                // Get is used to create a slice of the jobs to run which is passed to the doEvery ticker running each job on a schedule.
	Check(time.Time)           // Check is used to check the tasks tracking table to see if and when users need to be sent reminders.
	CheckCompliance(time.Time) // CheckCompliance exists to check those users who went beyond the deadline and were put into the overdue table while users can self remediate by messaging Cuebert, this is a failsafe to ensure that users are not left in jail.
	Diff(time.Time)            // Diff is used to compare what is in the database with "live" data from the task source.
	Poll(time.Time)            // Poll runs to check for custom reminders that users may have set for themselves.
	Populate(time.Time) error  // Populate is used to keep the content within the database, for the task at hand, up to date with information.
}

// Setup is used to setup the task.
type Setup interface {
	Setup(*Config)
}
