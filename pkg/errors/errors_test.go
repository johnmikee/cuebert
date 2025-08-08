package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError(t *testing.T) {
	t.Run("New creates error with code and message", func(t *testing.T) {
		err := New(ErrCodeValidation, "invalid input")

		assert.Equal(t, ErrCodeValidation, err.Code)
		assert.Equal(t, "invalid input", err.Message)
		assert.Nil(t, err.Err)
		assert.Contains(t, err.Error(), "[VALIDATION] invalid input")
		assert.NotEmpty(t, err.StackTrace)
	})

	t.Run("Wrap wraps existing error", func(t *testing.T) {
		originalErr := errors.New("connection refused")
		err := Wrap(originalErr, ErrCodeDBConnection, "failed to connect to database")

		assert.Equal(t, ErrCodeDBConnection, err.Code)
		assert.Equal(t, "failed to connect to database", err.Message)
		assert.Equal(t, originalErr, err.Err)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("Wrap returns nil for nil error", func(t *testing.T) {
		err := Wrap(nil, ErrCodeInternal, "should not create error")
		assert.Nil(t, err)
	})

	t.Run("Wrapf formats message", func(t *testing.T) {
		originalErr := errors.New("timeout")
		err := Wrapf(originalErr, ErrCodeTimeout, "operation %s timed out after %d seconds", "query", 30)

		assert.Equal(t, "operation query timed out after 30 seconds", err.Message)
	})

	t.Run("WithDetails adds context", func(t *testing.T) {
		err := New(ErrCodeValidation, "invalid user").
			WithDetails("user_id", "123").
			WithDetails("field", "email")

		assert.Equal(t, "123", err.Details["user_id"])
		assert.Equal(t, "email", err.Details["field"])
	})

	t.Run("Unwrap returns wrapped error", func(t *testing.T) {
		originalErr := errors.New("original")
		err := Wrap(originalErr, ErrCodeInternal, "wrapped")

		unwrapped := err.Unwrap()
		assert.Equal(t, originalErr, unwrapped)
	})
}

func TestErrorChecking(t *testing.T) {
	t.Run("Is checks error code", func(t *testing.T) {
		err := New(ErrCodeDBNotFound, "user not found")

		assert.True(t, Is(err, ErrCodeDBNotFound))
		assert.False(t, Is(err, ErrCodeDBQuery))
	})

	t.Run("Is works with wrapped errors", func(t *testing.T) {
		originalErr := New(ErrCodePermission, "access denied")
		wrappedErr := Wrap(originalErr, ErrCodeInternal, "operation failed")

		// Should preserve original code when wrapping with Internal
		assert.True(t, Is(wrappedErr, ErrCodePermission))
	})

	t.Run("As converts to AppError", func(t *testing.T) {
		err := New(ErrCodeSlackAPI, "slack error")

		appErr, ok := As(err)
		require.True(t, ok)
		assert.Equal(t, ErrCodeSlackAPI, appErr.Code)
	})

	t.Run("As returns false for non-AppError", func(t *testing.T) {
		err := errors.New("regular error")

		_, ok := As(err)
		assert.False(t, ok)
	})

	t.Run("IsCode checks specific code", func(t *testing.T) {
		err := New(ErrCodeConfig, "config error")

		assert.True(t, err.IsCode(ErrCodeConfig))
		assert.False(t, err.IsCode(ErrCodeDBQuery))
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("DBError detects connection errors", func(t *testing.T) {
		err := errors.New("connection refused")
		dbErr := DBError("select", err)

		assert.Equal(t, ErrCodeDBConnection, dbErr.Code)
		assert.Contains(t, dbErr.Message, "select")
	})

	t.Run("DBError detects not found errors", func(t *testing.T) {
		err := errors.New("no rows in result set")
		dbErr := DBError("find user", err)

		assert.Equal(t, ErrCodeDBNotFound, dbErr.Code)
	})

	t.Run("DBError detects duplicate errors", func(t *testing.T) {
		err := errors.New("unique constraint violation")
		dbErr := DBError("insert", err)

		assert.Equal(t, ErrCodeDBDuplicate, dbErr.Code)
	})

	t.Run("SlackError creates slack error", func(t *testing.T) {
		err := errors.New("invalid auth")
		slackErr := SlackError("send message", err)

		assert.Equal(t, ErrCodeSlackAPI, slackErr.Code)
		assert.Contains(t, slackErr.Message, "send message")
	})

	t.Run("ValidationError creates validation error", func(t *testing.T) {
		err := ValidationError("email", "invalid format")

		assert.Equal(t, ErrCodeValidation, err.Code)
		assert.Contains(t, err.Message, "email")
		assert.Contains(t, err.Message, "invalid format")
	})

	t.Run("NotFoundError creates not found error", func(t *testing.T) {
		err := NotFoundError("user")

		assert.Equal(t, ErrCodeNotFound, err.Code)
		assert.Contains(t, err.Message, "user not found")
	})

	t.Run("ConfigError with error", func(t *testing.T) {
		originalErr := errors.New("invalid json")
		err := ConfigError("database", originalErr)

		assert.Equal(t, ErrCodeConfig, err.Code)
		assert.Contains(t, err.Message, "database")
		assert.Equal(t, originalErr, err.Err)
	})

	t.Run("ConfigError without error", func(t *testing.T) {
		err := ConfigError("api_key", nil)

		assert.Equal(t, ErrCodeConfigMissing, err.Code)
		assert.Contains(t, err.Message, "api_key")
		assert.Nil(t, err.Err)
	})
}
