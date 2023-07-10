package main

import (
	"github.com/johnmikee/cuebert/cuebert/bot"
	"github.com/johnmikee/cuebert/cuebert/handlers"
	"github.com/johnmikee/cuebert/cuebert/method"
	"github.com/johnmikee/cuebert/cuebert/tables"
	"github.com/johnmikee/cuebert/db"
	"github.com/johnmikee/cuebert/idp"
	"github.com/johnmikee/cuebert/mdm"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Cuebert is a struct to hold config for the program
type Cuebert struct {
	bot           *bot.Bot
	config        *Config
	db            *db.DB
	flags         *Flags
	log           logger.Logger
	idp           idp.Provider
	mdm           mdm.Provider
	method        method.Actions
	tables        *tables.Config
	authUsers     []string
	testUsers     []string
	reloadSignal  chan struct{}
	statusChan    chan handlers.StatusMessage
	statusHandler *handlers.StatusHandler
	startSignal   chan struct{}
	stopSignal    chan struct{}
	isRunning     bool
}

// Config holds the sensitive values for the program
type Config struct {
	AdminGroupID      string `json:"admin_group_id"`
	DBAddress         string `json:"db_address"`
	DBName            string `json:"db_name"`
	DBPass            string `json:"db_pass"`
	DBPort            string `json:"db_port"`
	DBUser            string `json:"db_user"`
	IDPDomain         string `json:"idp_domain"`
	IDPToken          string `json:"idp_token"`
	IDPURL            string `json:"idp_url"`
	MDMKey            string `json:"mdm_key"`
	MDMURL            string `json:"mdm_url"`
	SlackAppToken     string `json:"slack_app_token"`
	SlackAlertChannel string `json:"slack_alert_channel"`
	SlackBotToken     string `json:"slack_bot_token"`
	SlackBotID        string `json:"slack_bot_id"`
}

// Flags holds the args for the program
//
// init is particularly useful when running in a container and you want
// be able to start, stop, and change the config from the bot itself.
type Flags struct {
	method                  string // ex: manager or time-bound. this sets the cadence for the flow of the program
	authUsers               string // comma separated list of users to perform authorized actions
	authUsersFromIDP        bool   // pull authorized users from the idp. if false use the auth-users flag
	checkInterval           int    // how often to check what cuebert messages need sending
	clearTables             bool   // clear all tables
	cutoffTime              string // cutoffTime will be the time access is revoked
	dailyReport             bool   // send a daily report to the slack channel
	deadline                string // the day the update is required
	defaultReminderInterval int    // how often to remind users to update their devices (time-bound only)
	deviceDiffInterval      int    // how often to check what devices we need to add/remove
	envType                 string // ex: dev, prod
	helpDocsURL             string // url to the help docs
	helpRepoURL             string // url to this repo for the help menu
	helpTicketURL           string // url to the help ticketing system
	idp                     string // ex: okta, onelogin
	init                    bool   // initialize the program and wait for input.
	logLevel                string // ex: debug, trace, info, warn, error
	logToFile               bool   // log to file defaults to false
	mdm                     string // ex: jamf, kandji
	pollInterval            int    // how often to poll for reminders
	requiredVers            string // ex: 13.1
	rebuildTablesOnFailure  bool   // rebuild tables on an abnormal exit.
	sendManagerMissing      bool   // send a message to the alert channel of missing managers
	serviceName             string // ex: cuebert
	tableNames              string // comma separated list of tables to clear
	testing                 bool   // run in testing mode
	testingEndTime          string // the hour the messaging should end
	testingStartTime        string // the hour the messaging should start
	testingUsers            string // comma separated list of users to test with
}

// TODO: this needs to log the bot flags via an interface
func (c *Cuebert) logFlags() {
	c.log.Trace().
		Str("authUsers", c.flags.authUsers).
		Bool("authUsersFromIDP", c.flags.authUsersFromIDP).
		Int("checkInterval", c.flags.checkInterval).
		Bool("clearTables", c.flags.clearTables).
		Str("cutoffTime", c.flags.cutoffTime).
		Bool("dailyReport", c.flags.dailyReport).
		Str("deadline", c.flags.deadline).
		Int("deviceDiffInterval", c.flags.deviceDiffInterval).
		Str("envType", c.flags.envType).
		Str("helpDocsURL", c.flags.helpDocsURL).
		Str("helpRepoURL", c.flags.helpRepoURL).
		Str("helpTicketURL", c.flags.helpTicketURL).
		Str("idp", c.flags.idp).
		Bool("init", c.flags.init).
		Str("tableNames", c.flags.tableNames).
		Str("logLevel", c.flags.logLevel).
		Bool("logToFile", c.flags.logToFile).
		Str("mdm", c.flags.mdm).
		Int("pollInterval", c.flags.pollInterval).
		Str("requiredVersion", c.flags.requiredVers).
		Bool("rebuildTablesOnFailure", c.flags.rebuildTablesOnFailure).
		Bool("sendManagerMissing", c.flags.sendManagerMissing).
		Str("serviceName", c.flags.serviceName).
		Str("tableNames", c.flags.tableNames).
		Bool("testing", c.flags.testing).
		Str("testingEndTime", c.flags.testingEndTime).
		Str("testingStartTime", c.flags.testingStartTime).
		Str("testingUsers", c.flags.testingUsers).
		Msg("current configuration")
}
