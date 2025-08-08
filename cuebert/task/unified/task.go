package unified

import (
	"context"
	"time"

	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// TaskType represents the type of task (compliance or update).
type TaskType string

const (
	TaskTypeCompliance TaskType = "compliance"
	TaskTypeUpdate     TaskType = "update"
)

// TaskStatus represents the current status of a task.
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusExcluded   TaskStatus = "excluded"
)

// Task represents a unified task interface for both compliance and update tasks.
type Task interface {
	// Identification
	GetID() string
	GetType() TaskType
	GetName() string
	GetDescription() string

	// User association
	GetUserID() string
	GetUserEmail() string

	// Status and tracking
	GetStatus() TaskStatus
	SetStatus(TaskStatus) error
	IsComplete() bool
	GetCompletedAt() *time.Time

	// Messaging
	NeedsFirstMessage() bool
	NeedsFinalMessage() bool
	NeedsReminder() bool

	// Execution
	Execute(ctx context.Context) error
	Validate(ctx context.Context) error

	// Metadata
	GetMetadata() map[string]interface{}
	SetMetadata(key string, value interface{})
}

// BaseTask provides common functionality for all task types.
type BaseTask struct {
	// Core fields
	ID          string   `json:"id" db:"id"`
	Type        TaskType `json:"type" db:"task_type"`
	Name        string   `json:"name" db:"task_name"`
	Description string   `json:"description" db:"description"`

	// User fields
	UserID       string `json:"user_id" db:"user_slack_id"`
	UserEmail    string `json:"user_email" db:"user_email"`
	UserFullName string `json:"user_full_name" db:"user_full_name"`

	// Status fields
	Status      TaskStatus `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`

	// Message tracking
	FirstMessageSent   bool       `json:"first_message_sent" db:"first_message_sent"`
	FirstMessageSentAt *time.Time `json:"first_message_sent_at,omitempty" db:"first_message_sent_at"`
	FinalMessageSent   bool       `json:"final_message_sent" db:"final_message_sent"`
	FinalMessageSentAt *time.Time `json:"final_message_sent_at,omitempty" db:"final_message_sent_at"`

	// Reminder fields
	ReminderEnabled bool       `json:"reminder_enabled" db:"reminder_enabled"`
	ReminderSentAt  *time.Time `json:"reminder_sent_at,omitempty" db:"reminder_sent_at"`
	NextReminderAt  *time.Time `json:"next_reminder_at,omitempty" db:"next_reminder_at"`

	// Metadata for task-specific data
	Metadata map[string]interface{} `json:"metadata" db:"metadata"`

	// Dependencies
	log logger.Logger
}

// GetID returns the task ID.
func (t *BaseTask) GetID() string {
	return t.ID
}

// GetType returns the task type.
func (t *BaseTask) GetType() TaskType {
	return t.Type
}

// GetName returns the task name.
func (t *BaseTask) GetName() string {
	return t.Name
}

// GetDescription returns the task description.
func (t *BaseTask) GetDescription() string {
	return t.Description
}

// GetUserID returns the user ID.
func (t *BaseTask) GetUserID() string {
	return t.UserID
}

// GetUserEmail returns the user email.
func (t *BaseTask) GetUserEmail() string {
	return t.UserEmail
}

// GetStatus returns the current task status.
func (t *BaseTask) GetStatus() TaskStatus {
	return t.Status
}

// SetStatus updates the task status.
func (t *BaseTask) SetStatus(status TaskStatus) error {
	oldStatus := t.Status
	t.Status = status
	t.UpdatedAt = time.Now()

	// Set completion time if moving to completed
	if status == TaskStatusCompleted && oldStatus != TaskStatusCompleted {
		now := time.Now()
		t.CompletedAt = &now
	}

	t.log.Info().
		Str("task_id", t.ID).
		Str("old_status", string(oldStatus)).
		Str("new_status", string(status)).
		Msg("task status updated")

	return nil
}

// IsComplete returns whether the task is complete.
func (t *BaseTask) IsComplete() bool {
	return t.Status == TaskStatusCompleted
}

// GetCompletedAt returns the completion time.
func (t *BaseTask) GetCompletedAt() *time.Time {
	return t.CompletedAt
}

// NeedsFirstMessage checks if the first message needs to be sent.
func (t *BaseTask) NeedsFirstMessage() bool {
	return !t.FirstMessageSent && !t.IsComplete() && t.Status != TaskStatusExcluded
}

// NeedsFinalMessage checks if the final message needs to be sent.
func (t *BaseTask) NeedsFinalMessage() bool {
	return t.FirstMessageSent && !t.FinalMessageSent && !t.IsComplete() && t.Status != TaskStatusExcluded
}

// NeedsReminder checks if a reminder needs to be sent.
func (t *BaseTask) NeedsReminder() bool {
	if !t.ReminderEnabled || t.IsComplete() || t.Status == TaskStatusExcluded {
		return false
	}

	if t.NextReminderAt == nil {
		return false
	}

	return time.Now().After(*t.NextReminderAt)
}

// GetMetadata returns the task metadata.
func (t *BaseTask) GetMetadata() map[string]interface{} {
	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	return t.Metadata
}

// SetMetadata sets a metadata value.
func (t *BaseTask) SetMetadata(key string, value interface{}) {
	if t.Metadata == nil {
		t.Metadata = make(map[string]interface{})
	}
	t.Metadata[key] = value
}

// Validate performs basic validation.
func (t *BaseTask) Validate(ctx context.Context) error {
	if t.ID == "" {
		return errors.ValidationError("task ID is required")
	}
	if t.UserID == "" {
		return errors.ValidationError("user ID is required")
	}
	if t.Name == "" {
		return errors.ValidationError("task name is required")
	}
	return nil
}
