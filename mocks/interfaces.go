package mocks

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/slack-go/slack"
)

// This file defines all the interfaces that need mocking for testing
// Using interfaces allows us to inject mock implementations during tests

// DatabaseClient defines the database operations interface.
type DatabaseClient interface {
	Execute(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
	Transaction(ctx context.Context, fn func(pgx.Tx) error) error
	Ping(ctx context.Context) error
	Close()
}

// SlackClient defines the Slack operations interface.
type SlackClient interface {
	SendMessage(channelID string, options ...slack.MsgOption) (string, string, string, error)
	SendMessageContext(ctx context.Context, channelID string, options ...slack.MsgOption) (string, string, string, error)
	OpenView(triggerID string, view slack.ModalViewRequest) (*slack.ViewResponse, error)
	UpdateView(viewID string, view slack.ModalViewRequest) (*slack.ViewResponse, error)
	GetUserInfo(userID string) (*slack.User, error)
	GetUsersInConversation(params *slack.GetUsersInConversationParameters) ([]string, string, error)
	OpenConversation(params *slack.OpenConversationParameters) (*slack.Channel, bool, bool, error)
	PostEphemeral(channelID, userID string, options ...slack.MsgOption) (string, error)
}

// MDMClient defines the MDM operations interface.
type MDMClient interface {
	GetDevice(ctx context.Context, serialNumber string) (Device, error)
	GetDeviceByUser(ctx context.Context, userEmail string) ([]Device, error)
	GetAllDevices(ctx context.Context) ([]Device, error)
	UpdateDevice(ctx context.Context, serialNumber string, updates map[string]interface{}) error
}

// IDPClient defines the IDP operations interface.
type IDPClient interface {
	GetUser(ctx context.Context, userID string) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetGroupMembers(ctx context.Context, groupID string) ([]User, error)
	GetUserGroups(ctx context.Context, userID string) ([]Group, error)
	UpdateUser(ctx context.Context, userID string, updates map[string]interface{}) error
}

// TaskScheduler defines the task scheduling interface.
type TaskScheduler interface {
	Schedule(ctx context.Context, task Task) error
	Cancel(ctx context.Context, taskID string) error
	List(ctx context.Context) ([]Task, error)
	Get(ctx context.Context, taskID string) (Task, error)
}

// NotificationService defines the notification interface.
type NotificationService interface {
	SendFirstMessage(ctx context.Context, userID string, task Task) error
	SendFinalMessage(ctx context.Context, userID string, task Task) error
	SendReminder(ctx context.Context, userID string, task Task, reminderTime time.Time) error
	SendManagerNotification(ctx context.Context, managerID string, userIDs []string, task Task) error
}

// ConfigService defines the configuration interface.
type ConfigService interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}) error
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetDuration(key string) time.Duration
	Reload() error
	Watch(key string, callback func(value interface{})) error
}

// MetricsCollector defines the metrics collection interface.
type MetricsCollector interface {
	RecordMessageSent(messageType string)
	RecordMessageAcknowledged(messageType string, duration time.Duration)
	RecordTaskCompleted(taskName string, duration time.Duration)
	RecordError(errorType string)
	GetStats() Stats
}

// Data structures used by interfaces

// Device represents a device from MDM.
type Device struct {
	SerialNumber string
	UserEmail    string
	OS           string
	OSVersion    string
	Platform     string
	LastCheckin  time.Time
	Compliant    bool
}

// User represents a user from IDP.
type User struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	ManagerID string
	Active    bool
	Groups    []string
}

// Group represents a group from IDP.
type Group struct {
	ID          string
	Name        string
	Description string
	Members     []string
}

// Task represents a scheduled task.
type Task struct {
	ID          string
	Name        string
	Type        string
	UserID      string
	ScheduledAt time.Time
	Deadline    time.Time
	Status      string
	Metadata    map[string]interface{}
}

// Stats represents metrics statistics.
type Stats struct {
	MessagesSent          int64
	MessagesAcknowledged  int64
	TasksCompleted        int64
	Errors                int64
	AverageResponseTime   time.Duration
	AverageCompletionTime time.Duration
}

// BotContext defines the bot interaction context interface.
type BotContext interface {
	Context() context.Context
	Event() interface{}
	ChannelID() string
	UserID() string
	Callback() *slack.InteractionCallback
}
