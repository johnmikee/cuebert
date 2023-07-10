package helpers

import (
	"testing"

	"github.com/slack-go/slack"
)

func TestExtractEmails(t *testing.T) {
	message := "Please send an email to john.doe@example.com"
	expected := "john.doe@example.com"
	email := ExtractEmails(message)
	if email != expected {
		t.Errorf("Expected email: %s, but got: %s", expected, email)
	}
}

func TestTzStringer(t *testing.T) {
	tz := int64(-300)
	expected := "-300"
	tzStr := TZStringer(tz)
	if tzStr != expected {
		t.Errorf("Expected tz string: %s, but got: %s", expected, tzStr)
	}
}

func TestStringBooler(t *testing.T) {
	boolStr := "true"
	expected := true
	b := StringBooler(boolStr)
	if b != expected {
		t.Errorf("Expected bool value: %t, but got: %t", expected, b)
	}
}

func TestYnToBool(t *testing.T) {
	ynStr := "Yes"
	expected := true
	boolVal := YNToBool(ynStr)
	if boolVal != expected {
		t.Errorf("Expected bool value: %t, but got: %t", expected, boolVal)
	}
}

func TestValToInt(t *testing.T) {
	numStr := "42"
	expected := 42
	num := ValToInt(numStr)
	if num != expected {
		t.Errorf("Expected integer value: %d, but got: %d", expected, num)
	}
}

func TestOptsToStrs(t *testing.T) {
	opts := []slack.OptionBlockObject{
		{Value: "Option 1"},
		{Value: "Option 2"},
		{Value: "Option 3"},
	}
	expected := "Option 1,Option 2,Option 3"
	str := OptsToStrs(opts)
	if str != expected {
		t.Errorf("Expected string: %s, but got: %s", expected, str)
	}
}

func TestUsersToStrs(t *testing.T) {
	users := []string{"user1", "user2", "user3"}
	expected := "user1,user2,user3"
	userStr := UsersToStrs(users)
	if userStr != expected {
		t.Errorf("Expected user string: %s, but got: %s", expected, userStr)
	}
}

func TestMain(t *testing.T) {
	// Run tests
	testCases := []struct {
		name     string
		testFunc func(*testing.T)
	}{
		{"TestExtractEmails", TestExtractEmails},
		{"TestTzStringer", TestTzStringer},
		{"TestStringBooler", TestStringBooler},
		{"TestYnToBool", TestYnToBool},
		{"TestValToInt", TestValToInt},
		{"TestOptsToStrs", TestOptsToStrs},
		{"TestUsersToStrs", TestUsersToStrs},
	}

	for _, tc := range testCases {
		t.Run(tc.name, tc.testFunc)
	}

	// Run other tests or teardown
	// ...

	// Exit with the appropriate exit code

}
