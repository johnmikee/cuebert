package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
	"github.com/zalando/go-keyring"
)

// Manager handles all configuration for the application.
type Manager struct {
	mu       sync.RWMutex
	config   *AppConfig
	log      logger.Logger
	envType  string
	watchers map[string][]func(interface{})
}

// AppConfig represents the complete application configuration.
type AppConfig struct {
	// Core settings
	Service ServiceConfig `json:"service" yaml:"service"`

	// Database configuration
	Database DatabaseConfig `json:"database" yaml:"database"`

	// Slack configuration
	Slack SlackConfig `json:"slack" yaml:"slack"`

	// Task configuration
	Tasks TaskConfig `json:"tasks" yaml:"tasks"`

	// MDM configuration
	MDM MDMConfig `json:"mdm" yaml:"mdm"`

	// IDP configuration
	IDP IDPConfig `json:"idp" yaml:"idp"`

	// Notification settings
	Notifications NotificationConfig `json:"notifications" yaml:"notifications"`

	// Runtime settings
	Runtime RuntimeConfig `json:"runtime" yaml:"runtime"`
}

// ServiceConfig contains service-level settings.
type ServiceConfig struct {
	Name        string `json:"name" yaml:"name"`
	Version     string `json:"version" yaml:"version"`
	Environment string `json:"environment" yaml:"environment"`
	LogLevel    string `json:"log_level" yaml:"log_level"`
	LogToFile   bool   `json:"log_to_file" yaml:"log_to_file"`
	Debug       bool   `json:"debug" yaml:"debug"`
}

// DatabaseConfig contains database settings.
type DatabaseConfig struct {
	Host             string        `json:"host" yaml:"host"`
	Port             int           `json:"port" yaml:"port"`
	Database         string        `json:"database" yaml:"database"`
	User             string        `json:"user" yaml:"user"`
	Password         string        `json:"password" yaml:"password" sensitive:"true"`
	SSLMode          string        `json:"ssl_mode" yaml:"ssl_mode"`
	MaxConnections   int32         `json:"max_connections" yaml:"max_connections"`
	MinConnections   int32         `json:"min_connections" yaml:"min_connections"`
	ConnectTimeout   time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	QueryTimeout     time.Duration `json:"query_timeout" yaml:"query_timeout"`
	MigrationsPath   string        `json:"migrations_path" yaml:"migrations_path"`
	ClearTables      bool          `json:"clear_tables" yaml:"clear_tables"`
	RebuildOnFailure bool          `json:"rebuild_on_failure" yaml:"rebuild_on_failure"`
}

// SlackConfig contains Slack settings.
type SlackConfig struct {
	BotToken        string   `json:"bot_token" yaml:"bot_token" sensitive:"true"`
	AppToken        string   `json:"app_token" yaml:"app_token" sensitive:"true"`
	SigningSecret   string   `json:"signing_secret" yaml:"signing_secret" sensitive:"true"`
	AdminChannel    string   `json:"admin_channel" yaml:"admin_channel"`
	AlertChannel    string   `json:"alert_channel" yaml:"alert_channel"`
	AuthorizedUsers []string `json:"authorized_users" yaml:"authorized_users"`
	Testing         bool     `json:"testing" yaml:"testing"`
	TestingUsers    []string `json:"testing_users" yaml:"testing_users"`
}

// TaskConfig contains task-related settings.
type TaskConfig struct {
	// Common task settings
	DeadlineDate  string        `json:"deadline_date" yaml:"deadline_date"`
	CutoffTime    string        `json:"cutoff_time" yaml:"cutoff_time"`
	CheckInterval time.Duration `json:"check_interval" yaml:"check_interval"`
	DiffInterval  time.Duration `json:"diff_interval" yaml:"diff_interval"`
	PollInterval  time.Duration `json:"poll_interval" yaml:"poll_interval"`

	// OS Update specific
	RequiredOS string `json:"required_os" yaml:"required_os"`

	// Compliance specific
	ComplianceTaskID string `json:"compliance_task_id" yaml:"compliance_task_id"`
	ComplianceSource string `json:"compliance_source" yaml:"compliance_source"`
}

// MDMConfig contains MDM settings.
type MDMConfig struct {
	Provider  string        `json:"provider" yaml:"provider"` // jamf or kandji
	BaseURL   string        `json:"base_url" yaml:"base_url"`
	APIKey    string        `json:"api_key" yaml:"api_key" sensitive:"true"`
	APISecret string        `json:"api_secret" yaml:"api_secret" sensitive:"true"`
	Username  string        `json:"username" yaml:"username"`
	Password  string        `json:"password" yaml:"password" sensitive:"true"`
	Timeout   time.Duration `json:"timeout" yaml:"timeout"`
}

// IDPConfig contains IDP settings.
type IDPConfig struct {
	Provider     string        `json:"provider" yaml:"provider"` // okta
	BaseURL      string        `json:"base_url" yaml:"base_url"`
	APIKey       string        `json:"api_key" yaml:"api_key" sensitive:"true"`
	AdminGroupID string        `json:"admin_group_id" yaml:"admin_group_id"`
	AuthFromIDP  bool          `json:"auth_from_idp" yaml:"auth_from_idp"`
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
}

// NotificationConfig contains notification settings.
type NotificationConfig struct {
	FirstMessageEnabled bool          `json:"first_message_enabled" yaml:"first_message_enabled"`
	FinalMessageEnabled bool          `json:"final_message_enabled" yaml:"final_message_enabled"`
	ReminderInterval    time.Duration `json:"reminder_interval" yaml:"reminder_interval"`
	SendManagerMissing  bool          `json:"send_manager_missing" yaml:"send_manager_missing"`
	DailyReport         bool          `json:"daily_report" yaml:"daily_report"`
	DailyReportTime     string        `json:"daily_report_time" yaml:"daily_report_time"`
	TestingStartTime    string        `json:"testing_start_time" yaml:"testing_start_time"`
	TestingEndTime      string        `json:"testing_end_time" yaml:"testing_end_time"`
}

// RuntimeConfig contains runtime settings.
type RuntimeConfig struct {
	Init          bool   `json:"init" yaml:"init"`
	Method        string `json:"method" yaml:"method"` // manager or device
	HelpDocsURL   string `json:"help_docs_url" yaml:"help_docs_url"`
	HelpRepoURL   string `json:"help_repo_url" yaml:"help_repo_url"`
	HelpTicketURL string `json:"help_ticket_url" yaml:"help_ticket_url"`
}

// New creates a new configuration manager.
func New(envType string, log logger.Logger) (*Manager, error) {
	manager := &Manager{
		envType:  envType,
		log:      logger.ChildLogger("config", &log),
		config:   &AppConfig{},
		watchers: make(map[string][]func(interface{})),
	}

	// Load configuration in order of precedence
	if err := manager.loadDefaults(); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfig, "failed to load defaults")
	}

	if err := manager.loadFromFile(); err != nil {
		// File is optional, log but don't fail
		manager.log.Debug().Err(err).Msg("no config file found")
	}

	if err := manager.loadFromEnv(); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfig, "failed to load env vars")
	}

	if envType == "dev" {
		if err := manager.loadFromKeychain(); err != nil {
			manager.log.Debug().Err(err).Msg("no keychain values found")
		}
	}

	// Validate configuration
	if err := manager.validate(); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeConfigInvalid, "configuration validation failed")
	}

	manager.log.Info().Str("env", envType).Msg("configuration loaded")

	return manager, nil
}

// loadDefaults sets default configuration values.
func (m *Manager) loadDefaults() error {
	m.config = &AppConfig{
		Service: ServiceConfig{
			Name:        "cuebert",
			Environment: m.envType,
			LogLevel:    "info",
			Debug:       false,
		},
		Database: DatabaseConfig{
			Host:           "localhost",
			Port:           5432,
			Database:       "cue",
			User:           "cue",
			SSLMode:        "prefer",
			MaxConnections: 25,
			MinConnections: 5,
			ConnectTimeout: 10 * time.Second,
			QueryTimeout:   30 * time.Second,
			ClearTables:    true,
		},
		Slack: SlackConfig{
			Testing: true,
		},
		Tasks: TaskConfig{
			CheckInterval: 30 * time.Minute,
			DiffInterval:  60 * time.Minute,
			PollInterval:  15 * time.Minute,
		},
		MDM: MDMConfig{
			Provider: "kandji",
			Timeout:  30 * time.Second,
		},
		IDP: IDPConfig{
			Provider:    "okta",
			AuthFromIDP: true,
			Timeout:     30 * time.Second,
		},
		Notifications: NotificationConfig{
			FirstMessageEnabled: true,
			FinalMessageEnabled: true,
			ReminderInterval:    60 * time.Minute,
			DailyReport:         false,
			TestingStartTime:    "11:00",
			TestingEndTime:      "17:00",
		},
		Runtime: RuntimeConfig{
			Init:          true,
			Method:        "manager",
			HelpDocsURL:   "https://help.megacorp.com/cuebert",
			HelpRepoURL:   "https://github.com/johnmikee/cuebert",
			HelpTicketURL: "https://tickets.megacorp.com/cuebert",
		},
	}

	return nil
}

// loadFromFile loads configuration from a JSON file.
func (m *Manager) loadFromFile() error {
	configPaths := []string{
		"config.json",
		"config/config.json",
		"/etc/cuebert/config.json",
		fmt.Sprintf("%s/.cuebert/config.json", os.Getenv("HOME")),
	}

	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(data, m.config); err != nil {
			return errors.Wrapf(err, errors.ErrCodeConfig, "failed to parse config file: %s", path)
		}

		m.log.Info().Str("path", path).Msg("loaded config from file")
		return nil
	}

	return fmt.Errorf("no config file found")
}

// loadFromEnv loads configuration from environment variables.
func (m *Manager) loadFromEnv() error {
	// Map environment variables to config fields
	envMappings := map[string]func(string){
		"CUEBERT_DB_HOST":         func(v string) { m.config.Database.Host = v },
		"CUEBERT_DB_PORT":         func(v string) { fmt.Sscanf(v, "%d", &m.config.Database.Port) },
		"CUEBERT_DB_NAME":         func(v string) { m.config.Database.Database = v },
		"CUEBERT_DB_USER":         func(v string) { m.config.Database.User = v },
		"CUEBERT_DB_PASSWORD":     func(v string) { m.config.Database.Password = v },
		"CUEBERT_SLACK_BOT_TOKEN": func(v string) { m.config.Slack.BotToken = v },
		"CUEBERT_SLACK_APP_TOKEN": func(v string) { m.config.Slack.AppToken = v },
		"CUEBERT_MDM_API_KEY":     func(v string) { m.config.MDM.APIKey = v },
		"CUEBERT_IDP_API_KEY":     func(v string) { m.config.IDP.APIKey = v },
		"CUEBERT_LOG_LEVEL":       func(v string) { m.config.Service.LogLevel = v },
		"CUEBERT_DEBUG":           func(v string) { m.config.Service.Debug = v == "true" },
		"CUEBERT_TESTING":         func(v string) { m.config.Slack.Testing = v == "true" },
		"CUEBERT_REQUIRED_OS":     func(v string) { m.config.Tasks.RequiredOS = v },
		"CUEBERT_DEADLINE_DATE":   func(v string) { m.config.Tasks.DeadlineDate = v },
	}

	for env, setter := range envMappings {
		if value := os.Getenv(env); value != "" {
			setter(value)
		}
	}

	// Handle comma-separated lists
	if authUsers := os.Getenv("CUEBERT_AUTH_USERS"); authUsers != "" {
		m.config.Slack.AuthorizedUsers = strings.Split(authUsers, ",")
	}

	if testingUsers := os.Getenv("CUEBERT_TESTING_USERS"); testingUsers != "" {
		m.config.Slack.TestingUsers = strings.Split(testingUsers, ",")
	}

	return nil
}

// loadFromKeychain loads sensitive values from system keychain (dev mode).
func (m *Manager) loadFromKeychain() error {
	service := "cuebert"

	keychainMappings := map[string]*string{
		"db_password":     &m.config.Database.Password,
		"slack_bot_token": &m.config.Slack.BotToken,
		"slack_app_token": &m.config.Slack.AppToken,
		"mdm_api_key":     &m.config.MDM.APIKey,
		"idp_api_key":     &m.config.IDP.APIKey,
	}

	for key, target := range keychainMappings {
		value, err := keyring.Get(service, key)
		if err == nil && value != "" {
			*target = value
			m.log.Debug().Str("key", key).Msg("loaded from keychain")
		}
	}

	return nil
}

// validate ensures all required configuration is present.
func (m *Manager) validate() error {
	// Validate database config
	if m.config.Database.Host == "" {
		return errors.ValidationError("database.host", "required")
	}
	if m.config.Database.Database == "" {
		return errors.ValidationError("database.database", "required")
	}

	// Validate Slack config (if not in testing mode)
	if !m.config.Slack.Testing {
		if m.config.Slack.BotToken == "" {
			return errors.ValidationError("slack.bot_token", "required when not testing")
		}
		if m.config.Slack.AppToken == "" {
			return errors.ValidationError("slack.app_token", "required when not testing")
		}
	}

	// Validate task config
	if m.config.Tasks.DeadlineDate != "" {
		// Validate date format
		if _, err := time.Parse("2006-01-02", m.config.Tasks.DeadlineDate); err != nil {
			return errors.ValidationError("tasks.deadline_date", "invalid date format (use YYYY-MM-DD)")
		}
	}

	return nil
}

// Get returns the current configuration.
func (m *Manager) Get() *AppConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent mutations
	configCopy := *m.config
	return &configCopy
}

// GetDatabase returns database configuration.
func (m *Manager) GetDatabase() DatabaseConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Database
}

// GetSlack returns Slack configuration.
func (m *Manager) GetSlack() SlackConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Slack
}

// GetTasks returns task configuration.
func (m *Manager) GetTasks() TaskConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.Tasks
}

// GetMDM returns MDM configuration.
func (m *Manager) GetMDM() MDMConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.MDM
}

// GetIDP returns IDP configuration.
func (m *Manager) GetIDP() IDPConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config.IDP
}

// Update updates configuration values at runtime.
func (m *Manager) Update(updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Apply updates (simplified - in production would use reflection)
	for key, value := range updates {
		switch key {
		case "log_level":
			if v, ok := value.(string); ok {
				m.config.Service.LogLevel = v
				m.notifyWatchers("service.log_level", v)
			}
		case "debug":
			if v, ok := value.(bool); ok {
				m.config.Service.Debug = v
				m.notifyWatchers("service.debug", v)
			}
		case "testing":
			if v, ok := value.(bool); ok {
				m.config.Slack.Testing = v
				m.notifyWatchers("slack.testing", v)
			}
		}
	}

	return nil
}

// Watch registers a callback for configuration changes.
func (m *Manager) Watch(key string, callback func(interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watchers[key] = append(m.watchers[key], callback)
}

// notifyWatchers notifies all watchers of a configuration change.
func (m *Manager) notifyWatchers(key string, value interface{}) {
	if callbacks, ok := m.watchers[key]; ok {
		for _, callback := range callbacks {
			go callback(value)
		}
	}
}

// Reload reloads configuration from all sources.
func (m *Manager) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldConfig := m.config

	// Reload configuration
	if err := m.loadDefaults(); err != nil {
		return err
	}
	if err := m.loadFromFile(); err != nil {
		m.log.Debug().Err(err).Msg("no config file found")
	}
	if err := m.loadFromEnv(); err != nil {
		return err
	}
	if m.envType == "dev" {
		if err := m.loadFromKeychain(); err != nil {
			m.log.Debug().Err(err).Msg("no keychain values found")
		}
	}

	// Validate new configuration
	if err := m.validate(); err != nil {
		m.config = oldConfig // Restore old config on validation failure
		return err
	}

	// Notify all watchers
	m.notifyWatchers("reload", m.config)

	m.log.Info().Msg("configuration reloaded")
	return nil
}

// String returns a string representation of the config (with sensitive values hidden).
func (m *Manager) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy with sensitive values masked
	safeConfig := *m.config
	safeConfig.Database.Password = "***"
	safeConfig.Slack.BotToken = "***"
	safeConfig.Slack.AppToken = "***"
	safeConfig.Slack.SigningSecret = "***"
	safeConfig.MDM.APIKey = "***"
	safeConfig.MDM.APISecret = "***"
	safeConfig.MDM.Password = "***"
	safeConfig.IDP.APIKey = "***"

	data, _ := json.MarshalIndent(safeConfig, "", "  ")
	return string(data)
}
