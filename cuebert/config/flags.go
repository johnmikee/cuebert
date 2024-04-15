package config

import "github.com/johnmikee/cuebert/pkg/logger"

// Flags holds the args for the program
type Flags struct {
	AuthorizedUsers         string   // Comma-separated list of users to perform authorized actions.
	AuthUsers               []string // Composed version of AuthorizedUsers.
	AuthUsersFromIDP        bool     // Pull authorized users from the IDP. If false, use the auth-users flag.
	CheckInterval           int      // How often to check what Cuebert messages need sending.
	CutoffTime              string   // CutoffTime will be the time access is revoked.
	Deadline                string   // The day the update is required.
	Debug                   bool     // Run the slacker client in Debug mode.
	DefaultReminderInterval int      // How often to remind users to update their devices (time-bound only).
	DiffInterval            int      // How often to check what devices we need to add/remove.
	EnvType                 string   // Ex: dev, prod.
	HelpDocsURL             string   // URL to the help docs.
	HelpRepoURL             string   // URL to this repo for the help menu.
	HelpTicketURL           string   // URL to the help ticketing system.
	IDP                     string   // Ex: Okta, OneLogin.
	Init                    bool     // Initialize the program and wait for input.
	LogLevel                string   // Ex: debug, trace, info, warn, error.
	LogToFile               bool     // Log to file. Defaults to false.
	MakeMigrations          bool     // Create the migrations and exit.
	MDM                     string   // Ex: Jamf, Kandji.
	PollInterval            int      // How often to poll for reminders.
	RequiredVers            string   // Ex: 13.1.
	ServiceName             string   // ex: Cuebert.
	Task                    string   // ex: os or compliance. This sets the task the program will perform.
	TaskImplementation      string   // ex: drata or lessonly
	TaskID                  string   // ex: 12345
	Testing                 bool     // Run in testing mode.
	TestingEndTime          string   // The hour the messaging should end.
	TestingStartTime        string   // The hour the messaging should start.
	TestingUsers            string   // Comma-separated list of users to test with.
	TestUsers               []string // Composed version of TestingUsers.
}

// LogFlags logs the flags. Used when the configuration is updated.
func LogFlags(log logger.Logger, flags *Flags) {
	log.Trace().
		Strs("authUsers", flags.AuthUsers).
		Bool("authUsersFromIDP", flags.AuthUsersFromIDP).
		Int("checkInterval", flags.CheckInterval).
		Str("cutoffTime", flags.CutoffTime).
		Str("deadline", flags.Deadline).
		Bool("debug", flags.Debug).
		Int("DiffInterval", flags.DiffInterval).
		Str("envType", flags.EnvType).
		Str("helpDocsURL", flags.HelpDocsURL).
		Str("helpRepoURL", flags.HelpRepoURL).
		Str("helpTicketURL", flags.HelpTicketURL).
		Str("idp", flags.IDP).
		Bool("init", flags.Init).
		Str("logLevel", flags.LogLevel).
		Bool("logToFile", flags.LogToFile).
		Str("mdm", flags.MDM).
		Int("pollInterval", flags.PollInterval).
		Str("task", flags.Task).
		Str("taskImplementation", flags.TaskImplementation).
		Str("taskID", flags.TaskID).
		Str("requiredVersion", flags.RequiredVers).
		Str("serviceName", flags.ServiceName).
		Bool("testing", flags.Testing).
		Str("testingEndTime", flags.TestingEndTime).
		Str("testingStartTime", flags.TestingStartTime).
		Strs("testingUsers", flags.TestUsers).
		Msg("current configuration")
}
