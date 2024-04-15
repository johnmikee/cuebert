package ops

import (
	"strings"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/slack-go/slack"
)

/*
Given that each item is optional we need to create a function for each type (bool, int, string, []string)
 that takes the selected option as a parameter and if that is empty return the default value
*/

// OptionalBool takes a slack.OptionBlockObject and a current bool and returns the selected value or the current value
func OptionalBool(opt slack.OptionBlockObject, current bool) bool {
	if opt.Value == "" {
		return current
	}
	return helpers.YNToBool(opt.Text.Text)
}

// OptionalInt takes a string and a current int and returns the selected value or the current value
func OptionalInt(opt string, current int) int {
	if opt == "" {
		return current
	}
	return helpers.ValToInt(strings.TrimSuffix(opt, "m"))
}

// OptionalStr takes a string and a current string and returns the selected value or the current value
func OptionalStr(opt, current string) string {
	if opt == "" {
		return current
	}
	return opt
}

// OptionalStrSelect takes a slack.OptionBlockObject and a current string and returns the selected value or the current value
func OptionalStrSelect(opt slack.OptionBlockObject, current string) string {
	if opt.Value == "" {
		return current
	}
	return opt.Text.Text
}

func OptionalStrs(opts, current []string) []string {
	if opts == nil {
		return current
	}
	return opts
}

func OptionalStrsToStr(opts []slack.OptionBlockObject, current string) string {
	if len(opts) == 0 {
		return current
	}

	strs := []string{}
	for _, o := range opts {
		strs = append(strs, o.Value)
	}
	return helpers.OptsToStrs(strs)
}
