package helpers

import (
	"fmt"
	"time"
)

// Date will return a time.Time object in UTC
func Date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// StringToTime takes a string and formats it against the layout return
// a time.Time object or nothing
func StringToTime(timeString string) (time.Time, error) {
	date, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

// UpdateTime will return the current time in UTC
func UpdateTime() time.Time {
	return time.Now().UTC()
}

func DateDiffFromToday(dateString string) (int, error) {
	dateFormat := "2006-01-02"

	// Parse the input date string into a time.Time object.
	inputDate, err := time.Parse(dateFormat, dateString)
	if err != nil {
		fmt.Println("Error parsing date:", err)
		return 0, err
	}

	// Get the current date and time.
	currentDate := time.Now()

	// Calculate the duration between the two dates.
	duration := currentDate.Sub(inputDate)

	// Extract the number of days from the duration.
	daysDifference := int(duration.Hours() / 24)

	return daysDifference, nil
}
