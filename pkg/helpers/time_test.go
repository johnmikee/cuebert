package helpers

import (
	"strings"
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	year := 2023
	month := 5
	day := 29
	expected := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	result := Date(year, month, day)

	if result != expected {
		t.Errorf("Date failed: expected '%s', got '%s'", expected, result)
	}
}

func TestStringToTime(t *testing.T) {
	// Test case: Valid time string
	timeString := "2023-06-19T10:30:00Z"
	expectedResult := time.Date(2023, time.June, 19, 10, 30, 0, 0, time.UTC)

	result, err := StringToTime(timeString)
	if err != nil {
		t.Errorf("Test case failed: unexpected error: %v", err)
	}
	if !result.Equal(expectedResult) {
		t.Errorf("Test case failed: expected %v, got %v", expectedResult, result)
	}

	// Test case: Invalid time string
	invalidTimeString := "2023-06-19T10:30:00"
	expectedErrorMessage := "cannot parse"
	if _, err := StringToTime(invalidTimeString); err == nil || !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Errorf("Test case failed: expected error containing \"%s\", got %v", expectedErrorMessage, err)
	}
}

func TestUpdateTime(t *testing.T) {
	expected := time.Now().UTC()
	result := UpdateTime()

	// Allow a small time difference for comparison due to precision limitations
	timeDiff := result.Sub(expected)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	// We consider a difference of up to 1 second acceptable
	acceptableDiff := time.Second
	if timeDiff > acceptableDiff {
		t.Errorf("UpdateTime failed: expected '%s', got '%s'", expected, result)
	}
}
