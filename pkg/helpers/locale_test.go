package helpers

import (
	"testing"
	"time"
)

func TestInRange(t *testing.T) {
	// Set the desired time zone for testing
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal("Error loading time zone:", err)
	}

	// Define test cases with inputs and expected outputs
	testCases := []struct {
		start    string
		end      string
		current  string
		expected bool
	}{
		{"09:00", "17:00", "2023-05-30 17:00", true},
		{"09:00", "17:00", "2023-05-30 08:59", false},
		{"09:00", "17:00", "2023-05-30 16:00", true},
		{"09:00", "17:00", "2023-05-30 12:00", true},
		{"09:00", "17:00", "2023-05-30 17:01", false},
		{"09:00", "17:00", "2023-05-30 00:00", false},
		{"09:00", "17:00", "2023-05-30 23:59", false},
		{"09:00", "09:00", "2023-05-30 09:00", true},
		{"09:00", "09:00", "2023-05-30 08:59", false},
		{"09:00", "09:00", "2023-05-30 09:01", false},
		{"17:00", "09:00", "2023-05-30 17:00", true},
		{"17:00", "09:00", "2023-05-30 08:59", true},
	}

	for _, tc := range testCases {
		// Parse the current time string in the specified time zone
		currentTime, err := time.ParseInLocation("2006-01-02 15:04", tc.current, location)
		if err != nil {
			t.Fatal("Error parsing current time:", err)
		}

		// Invoke the function and get the actual result
		actual := InRange(tc.start, tc.end, currentTime)

		// Compare the actual result with the expected result
		if actual != tc.expected {
			t.Errorf("For start=%s, end=%s, current=%s, expected=%t, but got %t",
				tc.start, tc.end, tc.current, tc.expected, actual)
		}
	}
}
