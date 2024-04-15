package ops

import (
	"fmt"
	"strings"
	"testing"
)

func TestFuzzyMatchNonOpt_NoMatch(t *testing.T) {
	opt := "apple"
	opts := []string{"banana", "orange", "mango"}

	expected := fmt.Sprintf("Sorry `%s` isnt a valid option.\n\nValid options are:\n%s",
		opt,
		strings.Join(opts, " \n"))

	result := FuzzyMatchNonOpt(opt, opts)

	if result != expected {
		t.Errorf("FuzzyMatchNonOpt failed: expected '%s', got '%s'", expected, result)
	}
}
