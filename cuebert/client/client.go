package client

import (
	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/handlers"

	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/slack-go/slack"
)

// Client holds the config for different parts of the program.
type Client struct {
	Bot           *bot.Bot
	Config        *config.Config
	DB            db.Client
	Flags         *config.Flags
	Log           logger.Logger
	IDP           idp.Provider
	MDM           mdm.Provider
	SC            *slack.Client
	Shared        *shared.Shared
	Task          task.Rabbit
	StatusChan    chan handlers.StatusMessage
	StatusHandler *handlers.StatusHandler
	IsRunning     bool
}
