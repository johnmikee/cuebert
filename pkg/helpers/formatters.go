package helpers

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/slack-go/slack"
)

// extractEmails extracts email addresses from the given message
// using a regular expression and returns them as a string.
func ExtractEmails(message string) string {
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	emails := emailRegex.FindString(message)
	return emails
}

// tzStringer converts a timezone offset value from int64 to string.
func TZStringer(tz int64) string {
	return strconv.FormatInt(tz, 10)
}

// stringBooler converts a string representation of a boolean to a bool value.
func StringBooler(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}

func BoolStringer(b bool) string {
	return strconv.FormatBool(b)
}

// ynToBool converts a "yes" or "no" string to a bool value.
func YNToBool(s string) bool {
	return strings.ToLower(s) == "yes"
}

// valToInt converts a string value to an integer.
func ValToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// optsToStrs converts a slice of slack.OptionBlockObject values to a string.
// It extracts the Value field from each option and returns them as a comma-separated string.
func OptsToStrs(opts []slack.OptionBlockObject) string {
	var strs []string
	for _, o := range opts {
		strs = append(strs, o.Value)
	}

	if len(strs) > 1 {
		return strings.Join(strs, ",")
	}

	return strs[0]
}

// usersToStrs converts a slice of user IDs to a string.
// If there are multiple user IDs, it returns them as a comma-separated string.
func UsersToStrs(users []string) string {
	if len(users) > 1 {
		return strings.Join(users, ",")
	}

	return users[0]
}
