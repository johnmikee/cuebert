package helpers

import (
	"fmt"
	"math"
	"time"
)

func AddOffset(offset int64, d, t string, loc *time.Location) (time.Time, error) {
	// the time string will be something like "2023-05-04 17:05:00" when joined
	ts := d + " " + t + ":00"

	if loc == nil {
		loc = GenLocation(offset)
	}

	// parse the user input time into a time.Time value in the user's local timezone
	userTime, err := time.ParseInLocation("2006-01-02 15:04:05", ts, loc)

	return userTime, err
}

func FutureDate(d, t string, offset int64) bool {
	var ok bool

	diff, err := LocaleDiff(offset, d, t)
	if err != nil {
		return false
	}
	switch {
	case diff.Abs().Hours() <= 24: // today
		ok = true
	case math.Signbit(diff.Hours()): // the diff is negative i.e in the past
		ok = false
	case diff.Abs().Hours() >= 24: // at least a day away
		ok = true
	default:
		return false
	}

	return ok
}

func GenLocation(offset int64) *time.Location {
	return time.FixedZone("", int(offset))
}

func InRange(s, e string, current time.Time) bool {
	layout := "2006-01-02 15:04"

	// Convert the start and end time strings to time values in the specified time zone
	start, err := time.ParseInLocation(
		layout,
		fmt.Sprintf("%s %s", current.Format("2006-01-02"), s),
		current.Location())
	if err != nil {
		return false
	}
	end, err := time.ParseInLocation(
		layout,
		fmt.Sprintf("%s %s", current.Format("2006-01-02"), e),
		current.Location())
	if err != nil {
		return false
	}

	switch {
	case start.Before(end):
		return !current.Before(start) && !current.After(end)
	case start.Equal(end):
		// Handle the case when start time equals end time
		return current.Equal(start)
	case !end.Equal(current):
		// Handle the case when start time is after end time (overnight range)
		return !current.Before(end) || !current.After(start)
	default:
		return current.Equal(end)
	}
}

func UserTime(offset int64, loc *time.Location) time.Time {
	if loc == nil {
		loc = GenLocation(offset)
	}

	userTime := time.Now().In(loc)

	return userTime
}

func LocaleDiff(offset int64, d, t string) (time.Duration, error) {
	loc := time.FixedZone("", int(offset))
	inputTime, err := AddOffset(offset, d, t, loc)
	if err != nil {
		return 0, fmt.Errorf("could not get userTime: %s", err.Error())
	}

	currentTime := UserTime(offset, loc)

	diff := inputTime.Sub(currentTime)

	return diff, nil
}

// getReminderDay returns the next Wednesday. this is used to send the reminder to
// the manager.
func GetReminderDay() time.Time {
	now := time.Now()
	weekday := int(now.Weekday())

	// Calculate the number of days to Wednesday (3)
	daysToWednesday := (3 - weekday + 7) % 7

	// Add the number of days to the current time
	wednesday := now.AddDate(0, 0, daysToWednesday)

	return wednesday
}
