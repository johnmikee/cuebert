package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode represents standardized error codes.
type ErrorCode string

const (
	// Database errors.
	ErrCodeDBConnection  ErrorCode = "DB_CONNECTION"
	ErrCodeDBQuery       ErrorCode = "DB_QUERY"
	ErrCodeDBTransaction ErrorCode = "DB_TRANSACTION"
	ErrCodeDBNotFound    ErrorCode = "DB_NOT_FOUND"
	ErrCodeDBDuplicate   ErrorCode = "DB_DUPLICATE"
	ErrCodeDBMigration   ErrorCode = "DB_MIGRATION"

	// Slack/Bot errors.
	ErrCodeSlackAPI     ErrorCode = "SLACK_API"
	ErrCodeSlackAuth    ErrorCode = "SLACK_AUTH"
	ErrCodeSlackMessage ErrorCode = "SLACK_MESSAGE"
	ErrCodeSlackModal   ErrorCode = "SLACK_MODAL"
	ErrCodeSlackUser    ErrorCode = "SLACK_USER"

	// External service errors.
	ErrCodeMDM  ErrorCode = "MDM_SERVICE"
	ErrCodeIDP  ErrorCode = "IDP_SERVICE"
	ErrCodeHTTP ErrorCode = "HTTP_REQUEST"

	// Configuration errors.
	ErrCodeConfig        ErrorCode = "CONFIG"
	ErrCodeConfigMissing ErrorCode = "CONFIG_MISSING"
	ErrCodeConfigInvalid ErrorCode = "CONFIG_INVALID"

	// Business logic errors.
	ErrCodeValidation    ErrorCode = "VALIDATION"
	ErrCodePermission    ErrorCode = "PERMISSION"
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	ErrCodeInvalidState  ErrorCode = "INVALID_STATE"

	// System errors.
	ErrCodeInternal  ErrorCode = "INTERNAL"
	ErrCodeTimeout   ErrorCode = "TIMEOUT"
	ErrCodeCancelled ErrorCode = "CANCELLED"
)

// AppError represents a standardized application error.
type AppError struct {
	Code       ErrorCode
	Message    string
	Err        error
	Details    map[string]interface{}
	StackTrace string
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error.
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds additional context details.
func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// IsCode checks if the error has a specific error code.
func (e *AppError) IsCode(code ErrorCode) bool {
	return e.Code == code
}

// New creates a new AppError.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StackTrace: getStackTrace(2),
	}
}

// Wrap wraps an existing error with additional context.
func Wrap(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}

	// If already an AppError, preserve the original code if the new one is generic
	if appErr, ok := err.(*AppError); ok && code == ErrCodeInternal {
		appErr.Message = fmt.Sprintf("%s: %s", message, appErr.Message)
		return appErr
	}

	return &AppError{
		Code:       code,
		Message:    message,
		Err:        err,
		StackTrace: getStackTrace(2),
	}
}

// Wrapf wraps an error with a formatted message.
func Wrapf(err error, code ErrorCode, format string, args ...interface{}) *AppError {
	return Wrap(err, code, fmt.Sprintf(format, args...))
}

// Is checks if an error matches a target error or error code.
func Is(err error, target interface{}) bool {
	switch t := target.(type) {
	case ErrorCode:
		var appErr *AppError
		if errors.As(err, &appErr) {
			return appErr.Code == t
		}
	case error:
		return errors.Is(err, t)
	case *AppError:
		var appErr *AppError
		if errors.As(err, &appErr) {
			return appErr.Code == t.Code
		}
	}
	return false
}

// As attempts to convert an error to AppError.
func As(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}

// getStackTrace captures the current stack trace.
func getStackTrace(skip int) string {
	var builder strings.Builder
	pcs := make([]uintptr, 10)
	n := runtime.Callers(skip+1, pcs)

	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}

		builder.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))

		if !more {
			break
		}
	}

	return builder.String()
}

// Helper functions for common error patterns

// DBError creates a database error.
func DBError(operation string, err error) *AppError {
	code := ErrCodeDBQuery
	if strings.Contains(strings.ToLower(err.Error()), "connection") {
		code = ErrCodeDBConnection
	} else if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(strings.ToLower(err.Error()), "no rows") {
		code = ErrCodeDBNotFound
	} else if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique constraint") {
		code = ErrCodeDBDuplicate
	}

	return Wrapf(err, code, "database operation failed: %s", operation)
}

// SlackError creates a Slack API error.
func SlackError(operation string, err error) *AppError {
	return Wrapf(err, ErrCodeSlackAPI, "slack operation failed: %s", operation)
}

// ValidationError creates a validation error.
func ValidationError(field, reason string) *AppError {
	return New(ErrCodeValidation, fmt.Sprintf("validation failed for %s: %s", field, reason))
}

// NotFoundError creates a not found error.
func NotFoundError(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

// ConfigError creates a configuration error.
func ConfigError(field string, err error) *AppError {
	if err != nil {
		return Wrapf(err, ErrCodeConfig, "configuration error for field: %s", field)
	}
	return New(ErrCodeConfigMissing, fmt.Sprintf("missing required configuration: %s", field))
}
