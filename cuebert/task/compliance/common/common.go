package common

import (
	"github.com/johnmikee/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/config"
	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tables"
	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/handlers"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/idp/client"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Common is a struct to hold common config for the task client.
type Common struct {
	bot    *bot.Bot
	config *config.Config
	flags  *config.Flags
	idp    idp.Provider
	log    logger.Logger
	tables *tables.Config
	status *handlers.StatusHandler
	shared *shared.Shared
}

// Config is a struct used to initialize the common config for the task client.
type Config struct {
	Bot    *bot.Bot
	Config *config.Config
	Flags  *config.Flags
	Log    logger.Logger
	Tables *tables.Config
	Status *handlers.StatusHandler
	Shared *shared.Shared
}

// New will return a new common config for the task client.
func New(u *Config) *Common {
	c := &Common{
		bot:    u.Bot,
		config: u.Config,
		idp: client.New(
			&client.IDP{
				IDP: idp.IDP(u.Flags.IDP),
				Config: idp.Config{
					Domain:        u.Config.Subdomain,
					URL:           u.Config.IDPURL,
					Token:         u.Config.IDPToken,
					Client:        nil,
					Log:           u.Log,
					IDPBlockGroup: u.Config.IDPBlockGroup,
				},
			},
		),
		flags:  u.Flags,
		log:    logger.ChildLogger("common", &u.Log),
		tables: u.Tables,
		status: u.Status,
		shared: u.Shared,
	}

	return c
}
