package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunReturnsErrorWhenReadPasswordFails(t *testing.T) {
	pr := stubPasswordReader{ReturnError: true}
	result, err := run(pr)
	assert.Error(t, err)
	assert.Equal(t, errors.New("stubbed error"), err)
	assert.Equal(t, "", result)
}

func TestRunReturnsPasswordInput(t *testing.T) {
	pr := stubPasswordReader{Password: "password"}
	result, err := run(pr)
	assert.NoError(t, err)
	assert.Equal(t, "password", result)
}
