package helpers

import "time"

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
