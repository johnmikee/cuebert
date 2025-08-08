package unified

import (
	"context"
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/pkg/errors"
	"github.com/johnmikee/cuebert/pkg/logger"
)

// UpdateTask represents an OS or software update task
type UpdateTask struct {
	BaseTask

	// Update-specific fields
	DeviceID       string    `json:"device_id" db:"device_id"`
	DeviceSerial   string    `json:"device_serial" db:"device_serial"`
	DeviceOS       string    `json:"device_os" db:"device_os"`
	CurrentVersion string    `json:"current_version" db:"current_version"`
	TargetVersion  string    `json:"target_version" db:"target_version"`
	UpdateType     string    `json:"update_type" db:"update_type"` // os, security, app
	Deadline       time.Time `json:"deadline" db:"deadline"`

	// Update checker function
	checkUpdate func(ctx context.Context, deviceID string) (bool, string, error)
}

// NewUpdateTask creates a new update task
func NewUpdateTask(
	id, userID, userEmail, userFullName string,
	deviceID, deviceSerial, deviceOS, currentVersion, targetVersion string,
	deadline time.Time,
	log logger.Logger,
) *UpdateTask {
	taskName := fmt.Sprintf("%s Update", deviceOS)
	description := fmt.Sprintf("Update %s from %s to %s", deviceOS, currentVersion, targetVersion)

	return &UpdateTask{
		BaseTask: BaseTask{
			ID:           id,
			Type:         TaskTypeUpdate,
			Name:         taskName,
			Description:  description,
			UserID:       userID,
			UserEmail:    userEmail,
			UserFullName: userFullName,
			Status:       TaskStatusPending,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			log:          log,
		},
		DeviceID:       deviceID,
		DeviceSerial:   deviceSerial,
		DeviceOS:       deviceOS,
		CurrentVersion: currentVersion,
		TargetVersion:  targetVersion,
		UpdateType:     "os",
		Deadline:       deadline,
	}
}

// Execute runs the update check
func (u *UpdateTask) Execute(ctx context.Context) error {
	u.log.Info().
		Str("task_id", u.ID).
		Str("device_id", u.DeviceID).
		Str("target_version", u.TargetVersion).
		Msg("executing update task")

	// Update status to in progress
	if err := u.SetStatus(TaskStatusInProgress); err != nil {
		return errors.InternalError("update task status", err)
	}

	// Check if device has been updated
	if u.checkUpdate != nil {
		updated, currentVersion, err := u.checkUpdate(ctx, u.DeviceID)
		if err != nil {
			u.log.Error().
				Err(err).
				Str("task_id", u.ID).
				Str("device_id", u.DeviceID).
				Msg("update check failed")
			return errors.InternalError("check device update", err)
		}

		// Update current version
		u.CurrentVersion = currentVersion

		if updated {
			if err := u.SetStatus(TaskStatusCompleted); err != nil {
				return errors.InternalError("update task status", err)
			}
			u.log.Info().
				Str("task_id", u.ID).
				Str("device_id", u.DeviceID).
				Str("new_version", currentVersion).
				Msg("device has been updated")
		} else {
			// Still pending - device needs update
			if err := u.SetStatus(TaskStatusPending); err != nil {
				return errors.InternalError("update task status", err)
			}
			u.log.Debug().
				Str("task_id", u.ID).
				Str("device_id", u.DeviceID).
				Str("current_version", currentVersion).
				Str("target_version", u.TargetVersion).
				Msg("device still needs update")
		}
	}

	return nil
}

// Validate checks if the update task is valid
func (u *UpdateTask) Validate(ctx context.Context) error {
	// Run base validation
	if err := u.BaseTask.Validate(ctx); err != nil {
		return err
	}

	// Update-specific validation
	if u.DeviceID == "" {
		return errors.ValidationError("device ID is required")
	}

	if u.DeviceSerial == "" {
		return errors.ValidationError("device serial is required")
	}

	if u.TargetVersion == "" {
		return errors.ValidationError("target version is required")
	}

	if u.Deadline.IsZero() {
		return errors.ValidationError("deadline is required")
	}

	return nil
}

// SetUpdateChecker sets the update checking function
func (u *UpdateTask) SetUpdateChecker(checker func(ctx context.Context, deviceID string) (bool, string, error)) {
	u.checkUpdate = checker
}

// GetDeadline returns the update deadline
func (u *UpdateTask) GetDeadline() time.Time {
	return u.Deadline
}

// IsOverdue checks if the update is overdue
func (u *UpdateTask) IsOverdue() bool {
	return !u.IsComplete() && time.Now().After(u.Deadline)
}

// GetDaysUntilDeadline returns the number of days until the deadline
func (u *UpdateTask) GetDaysUntilDeadline() int {
	if u.IsComplete() {
		return 0
	}

	duration := u.Deadline.Sub(time.Now())
	days := int(duration.Hours() / 24)

	if days < 0 {
		return 0
	}

	return days
}

// NeedsUpdate checks if the device still needs the update
func (u *UpdateTask) NeedsUpdate() bool {
	return u.CurrentVersion != u.TargetVersion && !u.IsComplete()
}

// GetUpdateMetadata returns update-specific metadata
func (u *UpdateTask) GetUpdateMetadata() map[string]interface{} {
	metadata := u.GetMetadata()
	metadata["device_id"] = u.DeviceID
	metadata["device_serial"] = u.DeviceSerial
	metadata["device_os"] = u.DeviceOS
	metadata["current_version"] = u.CurrentVersion
	metadata["target_version"] = u.TargetVersion
	metadata["update_type"] = u.UpdateType
	metadata["deadline"] = u.Deadline
	metadata["is_overdue"] = u.IsOverdue()
	metadata["days_until_deadline"] = u.GetDaysUntilDeadline()
	metadata["needs_update"] = u.NeedsUpdate()

	return metadata
}
