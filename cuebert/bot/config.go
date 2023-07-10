package bot

type Cfg struct {
	authUsers          []string // comma separated list of users to perform authorized actions
	authUsersFromIDP   bool     // pull authorized users from the idp. if false use the auth-users flag
	checkInterval      int      // how often to check what cuebert messages need sending
	clearTables        bool     // clear all tables
	cutoffTime         string   // cutoffTime will be the time access is revoked
	deadline           string   // the day the update is required
	deviceDiffInterval int      // how often to check what devices we need to add/remove
	helpDocsURL        string   // url to the help docs
	helpRepoURL        string   // url to this repo for the help menu
	helpTicketURL      string   // url to the help ticketing system
	logLevel           string   // ex: debug
	logToFile          bool     // log to file defaults to false
	pollInterval       int      // how often to poll for reminders
	requiredVers       string   // ex: 13.1
	slackAlertChannel  string   // the slack channel to send alerts to
	slackBotID         string   // the slack bot id
	tableNames         string   // comma separated list of tables to clear
	testing            bool     // run in testing mode
	testingEndTime     string   // the hour the messaging should end
	testingStartTime   string   // the hour the messaging should start
	testUsers          []string // comma separated list of users to test with
}

type Option func(*Cfg)

func CfgSetter(opts ...Option) *Cfg {
	cfg := &Cfg{}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func (b *Bot) UpdateCfg(opts ...Option) {
	for _, opt := range opts {
		opt(b.cfg)
	}
}

func WithAuthUsers(users []string) Option {
	return func(cfg *Cfg) {
		cfg.authUsers = users
	}
}

func WithAuthUsersFromIDP(fromIDP bool) Option {
	return func(cfg *Cfg) {
		cfg.authUsersFromIDP = fromIDP
	}
}

func WithCheckInterval(interval int) Option {
	return func(cfg *Cfg) {
		cfg.checkInterval = interval
	}
}

func WithClearTables(clear bool) Option {
	return func(cfg *Cfg) {
		cfg.clearTables = clear
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

func WithDeviceDiffInterval(interval int) Option {
	return func(cfg *Cfg) {
		cfg.deviceDiffInterval = interval
	}
}

func WithHelpDocsURL(url string) Option {
	return func(cfg *Cfg) {
		cfg.helpDocsURL = url
	}
}

func WithHelpRepoURL(url string) Option {
	return func(cfg *Cfg) {
		cfg.helpRepoURL = url
	}
}

func WithHelpTicketURL(url string) Option {
	return func(cfg *Cfg) {
		cfg.helpTicketURL = url
	}
}

func WithLogLevel(level string) Option {
	return func(cfg *Cfg) {
		cfg.logLevel = level
	}
}

func WithLogToFile(logToFile bool) Option {
	return func(cfg *Cfg) {
		cfg.logToFile = logToFile
	}
}

func WithPollInterval(interval int) Option {
	return func(cfg *Cfg) {
		cfg.pollInterval = interval
	}
}

func WithRequiredVers(vers string) Option {
	return func(cfg *Cfg) {
		cfg.requiredVers = vers
	}
}

func WithSlackAlertChannel(channel string) Option {
	return func(cfg *Cfg) {
		cfg.slackAlertChannel = channel
	}
}

func WithSlackBotID(id string) Option {
	return func(cfg *Cfg) {
		cfg.slackBotID = id
	}
}

func WithTableNames(names string) Option {
	return func(cfg *Cfg) {
		cfg.tableNames = names
	}
}

func WithTesting(testing bool) Option {
	return func(cfg *Cfg) {
		cfg.testing = testing
	}
}

func WithTestingEndTime(time string) Option {
	return func(cfg *Cfg) {
		cfg.testingEndTime = time
	}
}

func WithTestingStartTime(time string) Option {
	return func(cfg *Cfg) {
		cfg.testingStartTime = time
	}
}

func WithTestUsers(users []string) Option {
	return func(cfg *Cfg) {
		cfg.testUsers = users
	}
}
