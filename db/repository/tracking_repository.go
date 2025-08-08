package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// Tracking represents a unified tracking entity for both compliance and updates.
type Tracking struct {
	// Common fields
	UserServiceID string `db:"user_service_id" json:"user_service_id"`
	UserSlackID   string `db:"user_slack_id" json:"user_slack_id"`
	UserEmail     string `db:"user_email" json:"user_email"`
	UserFullName  string `db:"user_full_name" json:"user_full_name"`
	TaskName      string `db:"task_name" json:"task_name"`
	TaskComplete  bool   `db:"task_complete" json:"task_complete"`

	// Message tracking
	FirstACK            bool      `db:"first_ack" json:"first_ack"`
	FirstACKTime        time.Time `db:"first_ack_time" json:"first_ack_time"`
	FirstMessageSent    bool      `db:"first_message_sent" json:"first_message_sent"`
	FirstMessageSentAt  time.Time `db:"first_message_sent_at" json:"first_message_sent_at"`
	FirstMessageWaiting bool      `db:"first_message_waiting" json:"first_message_waiting"`
	FinalMessageSent    bool      `db:"final_message_sent" json:"final_message_sent"`
	FinalMessageSentAt  time.Time `db:"final_message_sent_at" json:"final_message_sent_at"`

	// Reminder fields
	DelayAt          bool   `db:"delay_at" json:"delay_at"`
	DelayDate        string `db:"delay_date" json:"delay_date"`
	DelayTime        string `db:"delay_time" json:"delay_time"`
	ReminderSent     bool   `db:"reminder_sent" json:"reminder_sent"`
	ReminderInterval int    `db:"reminder_interval" json:"reminder_interval"`
	ReminderWaiting  bool   `db:"reminder_waiting" json:"reminder_waiting"`

	// Metadata
	TZOffset  int       `db:"tz_offset" json:"tz_offset"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// Optional fields (for update tracking)
	SerialNumber string `db:"serial_number,omitempty" json:"serial_number,omitempty"`
	DeviceOS     string `db:"device_os,omitempty" json:"device_os,omitempty"`
}

// TrackingType represents the type of tracking.
type TrackingType string

const (
	TrackingTypeCompliance TrackingType = "compliance"
	TrackingTypeUpdate     TrackingType = "update"
)

// TrackingRepository extends the base repository with tracking-specific methods.
type TrackingRepository interface {
	Repository[Tracking]

	// User-specific queries
	FindByUserSlackID(ctx context.Context, slackID string) ([]Tracking, error)
	FindByUserEmail(ctx context.Context, email string) ([]Tracking, error)

	// Task-specific queries
	FindByTaskName(ctx context.Context, taskName string) ([]Tracking, error)
	FindIncompleteByTask(ctx context.Context, taskName string) ([]Tracking, error)
	FindCompleteByTask(ctx context.Context, taskName string) ([]Tracking, error)

	// Message status queries
	FindUsersNeedingFirstMessage(ctx context.Context, taskName string) ([]Tracking, error)
	FindUsersNeedingFinalMessage(ctx context.Context, taskName string) ([]Tracking, error)
	FindUsersWithPendingReminders(ctx context.Context) ([]Tracking, error)

	// Update operations
	MarkFirstMessageSent(ctx context.Context, slackID string, taskName string) error
	MarkFinalMessageSent(ctx context.Context, slackID string, taskName string) error
	MarkTaskComplete(ctx context.Context, slackID string, taskName string) error
	UpdateReminderStatus(ctx context.Context, slackID string, taskName string, sent bool) error

	// Bulk operations
	MarkMultipleComplete(ctx context.Context, slackIDs []string, taskName string) (int64, error)
	ResetMessageStatus(ctx context.Context, taskName string) (int64, error)
}

type trackingRepository struct {
	Repository[Tracking]
	trackingType TrackingType
	tableName    string
}

// NewTrackingRepository creates a new tracking repository.
func NewTrackingRepository(db *pgxpool.Pool, log logger.Logger, trackingType TrackingType) TrackingRepository {
	tableName := "compliance_tracking"
	if trackingType == TrackingTypeUpdate {
		tableName = "bot_results"
	}

	base := New[Tracking](Config{
		DB:         db,
		Log:        log,
		Table:      tableName,
		PrimaryKey: "user_slack_id",
	})

	return &trackingRepository{
		Repository:   base,
		trackingType: trackingType,
		tableName:    tableName,
	}
}

// FindByUserSlackID finds tracking records by user Slack ID.
func (r *trackingRepository) FindByUserSlackID(ctx context.Context, slackID string) ([]Tracking, error) {
	return r.FindAll(ctx, Conditions{"user_slack_id": slackID})
}

// FindByUserEmail finds tracking records by user email.
func (r *trackingRepository) FindByUserEmail(ctx context.Context, email string) ([]Tracking, error) {
	return r.FindAll(ctx, Conditions{"user_email": email})
}

// FindByTaskName finds all tracking records for a specific task.
func (r *trackingRepository) FindByTaskName(ctx context.Context, taskName string) ([]Tracking, error) {
	return r.FindAll(ctx, Conditions{"task_name": taskName})
}

// FindIncompleteByTask finds incomplete tasks.
func (r *trackingRepository) FindIncompleteByTask(ctx context.Context, taskName string) ([]Tracking, error) {
	return r.FindAll(ctx, Conditions{
		"task_name":     taskName,
		"task_complete": false,
	})
}

// FindCompleteByTask finds completed tasks.
func (r *trackingRepository) FindCompleteByTask(ctx context.Context, taskName string) ([]Tracking, error) {
	return r.FindAll(ctx, Conditions{
		"task_name":     taskName,
		"task_complete": true,
	})
}

// FindUsersNeedingFirstMessage finds users who haven't received their first message.
func (r *trackingRepository) FindUsersNeedingFirstMessage(ctx context.Context, taskName string) ([]Tracking, error) {
	return r.FindAll(ctx, Conditions{
		"task_name":          taskName,
		"first_message_sent": false,
		"task_complete":      false,
	})
}

// FindUsersNeedingFinalMessage finds users who need their final message.
func (r *trackingRepository) FindUsersNeedingFinalMessage(ctx context.Context, taskName string) ([]Tracking, error) {
	return r.FindAll(ctx, Conditions{
		"task_name":          taskName,
		"final_message_sent": false,
		"first_message_sent": true,
		"task_complete":      false,
	})
}

// FindUsersWithPendingReminders finds users with pending reminders.
func (r *trackingRepository) FindUsersWithPendingReminders(ctx context.Context) ([]Tracking, error) {
	return r.FindAll(ctx, Conditions{
		"reminder_waiting": true,
		"task_complete":    false,
	})
}

// MarkFirstMessageSent marks the first message as sent.
func (r *trackingRepository) MarkFirstMessageSent(ctx context.Context, slackID string, taskName string) error {
	conditions := Conditions{
		"user_slack_id": slackID,
		"task_name":     taskName,
	}

	updates := Updates{
		"first_message_sent":    true,
		"first_message_sent_at": time.Now(),
	}

	affected, err := r.UpdateWhere(ctx, conditions, updates)
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.NotFoundError("tracking record")
	}

	return nil
}

// MarkFinalMessageSent marks the final message as sent.
func (r *trackingRepository) MarkFinalMessageSent(ctx context.Context, slackID string, taskName string) error {
	conditions := Conditions{
		"user_slack_id": slackID,
		"task_name":     taskName,
	}

	updates := Updates{
		"final_message_sent":    true,
		"final_message_sent_at": time.Now(),
	}

	affected, err := r.UpdateWhere(ctx, conditions, updates)
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.NotFoundError("tracking record")
	}

	return nil
}

// MarkTaskComplete marks a task as complete.
func (r *trackingRepository) MarkTaskComplete(ctx context.Context, slackID string, taskName string) error {
	conditions := Conditions{
		"user_slack_id": slackID,
		"task_name":     taskName,
	}

	updates := Updates{
		"task_complete": true,
	}

	affected, err := r.UpdateWhere(ctx, conditions, updates)
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.NotFoundError("tracking record")
	}

	return nil
}

// UpdateReminderStatus updates the reminder status.
func (r *trackingRepository) UpdateReminderStatus(ctx context.Context, slackID string, taskName string, sent bool) error {
	conditions := Conditions{
		"user_slack_id": slackID,
		"task_name":     taskName,
	}

	updates := Updates{
		"reminder_sent":    sent,
		"reminder_waiting": !sent,
	}

	affected, err := r.UpdateWhere(ctx, conditions, updates)
	if err != nil {
		return err
	}

	if affected == 0 {
		return errors.NotFoundError("tracking record")
	}

	return nil
}

// MarkMultipleComplete marks multiple users' tasks as complete.
func (r *trackingRepository) MarkMultipleComplete(ctx context.Context, slackIDs []string, taskName string) (int64, error) {
	// This would need a more complex query builder to handle IN clause
	// For now, iterate through IDs
	var totalAffected int64

	for _, slackID := range slackIDs {
		err := r.MarkTaskComplete(ctx, slackID, taskName)
		if err != nil {
			// Log error but continue
			continue
		}
		totalAffected++
	}

	return totalAffected, nil
}

// ResetMessageStatus resets all message statuses for a task.
func (r *trackingRepository) ResetMessageStatus(ctx context.Context, taskName string) (int64, error) {
	conditions := Conditions{
		"task_name": taskName,
	}

	updates := Updates{
		"first_message_sent":    false,
		"first_message_sent_at": nil,
		"final_message_sent":    false,
		"final_message_sent_at": nil,
		"reminder_sent":         false,
		"reminder_waiting":      false,
	}

	return r.UpdateWhere(ctx, conditions, updates)
}
