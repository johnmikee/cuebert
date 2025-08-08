package unified

import (
	"context"
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// ComplianceTask represents a compliance training or certification task.
type ComplianceTask struct {
	BaseTask

	// Compliance-specific fields
	TrainingProvider string    `json:"training_provider" db:"training_provider"`
	TrainingURL      string    `json:"training_url" db:"training_url"`
	DueDate          time.Time `json:"due_date" db:"due_date"`
	CertificateID    string    `json:"certificate_id,omitempty" db:"certificate_id"`

	// Compliance checker function
	checkCompliance func(ctx context.Context, userID string) (bool, error)
}

// NewComplianceTask creates a new compliance task.
func NewComplianceTask(
	id, userID, userEmail, userFullName, taskName string,
	provider string,
	dueDate time.Time,
	log logger.Logger,
) *ComplianceTask {
	return &ComplianceTask{
		BaseTask: BaseTask{
			ID:           id,
			Type:         TaskTypeCompliance,
			Name:         taskName,
			Description:  fmt.Sprintf("Complete %s training", provider),
			UserID:       userID,
			UserEmail:    userEmail,
			UserFullName: userFullName,
			Status:       TaskStatusPending,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			log:          log,
		},
		TrainingProvider: provider,
		DueDate:          dueDate,
	}
}

// Execute runs the compliance check.
func (c *ComplianceTask) Execute(ctx context.Context) error {
	c.log.Info().
		Str("task_id", c.ID).
		Str("user_id", c.UserID).
		Str("provider", c.TrainingProvider).
		Msg("executing compliance task")

	// Update status to in progress
	if err := c.SetStatus(TaskStatusInProgress); err != nil {
		return errors.InternalError("update task status", err)
	}

	// Check if user is compliant
	if c.checkCompliance != nil {
		compliant, err := c.checkCompliance(ctx, c.UserID)
		if err != nil {
			c.log.Error().
				Err(err).
				Str("task_id", c.ID).
				Str("user_id", c.UserID).
				Msg("compliance check failed")
			return errors.InternalError("check compliance", err)
		}

		if compliant {
			if err := c.SetStatus(TaskStatusCompleted); err != nil {
				return errors.InternalError("update task status", err)
			}
			c.log.Info().
				Str("task_id", c.ID).
				Str("user_id", c.UserID).
				Msg("user is compliant")
		} else {
			// Still pending - user needs to complete training
			if err := c.SetStatus(TaskStatusPending); err != nil {
				return errors.InternalError("update task status", err)
			}
		}
	}

	return nil
}

// Validate checks if the compliance task is valid.
func (c *ComplianceTask) Validate(ctx context.Context) error {
	// Run base validation
	if err := c.BaseTask.Validate(ctx); err != nil {
		return err
	}

	// Compliance-specific validation
	if c.TrainingProvider == "" {
		return errors.ValidationError("training provider is required")
	}

	if c.DueDate.IsZero() {
		return errors.ValidationError("due date is required")
	}

	if c.DueDate.Before(time.Now()) {
		return errors.ValidationError("due date cannot be in the past")
	}

	return nil
}

// SetComplianceChecker sets the compliance checking function.
func (c *ComplianceTask) SetComplianceChecker(checker func(ctx context.Context, userID string) (bool, error)) {
	c.checkCompliance = checker
}

// GetDueDate returns the task due date.
func (c *ComplianceTask) GetDueDate() time.Time {
	return c.DueDate
}

// IsOverdue checks if the task is overdue.
func (c *ComplianceTask) IsOverdue() bool {
	return !c.IsComplete() && time.Now().After(c.DueDate)
}

// GetDaysUntilDue returns the number of days until the task is due.
func (c *ComplianceTask) GetDaysUntilDue() int {
	if c.IsComplete() {
		return 0
	}

	duration := c.DueDate.Sub(time.Now())
	days := int(duration.Hours() / 24)

	if days < 0 {
		return 0
	}

	return days
}

// GetComplianceMetadata returns compliance-specific metadata.
func (c *ComplianceTask) GetComplianceMetadata() map[string]interface{} {
	metadata := c.GetMetadata()
	metadata["training_provider"] = c.TrainingProvider
	metadata["training_url"] = c.TrainingURL
	metadata["due_date"] = c.DueDate
	metadata["is_overdue"] = c.IsOverdue()
	metadata["days_until_due"] = c.GetDaysUntilDue()

	if c.CertificateID != "" {
		metadata["certificate_id"] = c.CertificateID
	}

	return metadata
}
