package ops

import (
	"fmt"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// FuzzyMatchOpt takes a string and a list of options and returns the closest match
func FuzzyMatchNonOpt(opt string, opts []string) string {
	maybeMatch := fuzzy.RankFind(opt, opts)
	var msg string

	if len(maybeMatch) > 0 {
		msg = fmt.Sprintf("Sorry `%s` isnt a valid option. Did you mean `%s`?\n\nValid options are:\n%s",
			opt,
			maybeMatch[0].Target,
			strings.Join(opts, " \n"))
	} else {
		msg = fmt.Sprintf("Sorry `%s` isnt a valid option.\n\nValid options are:\n%s",
			opt,
			strings.Join(opts, " \n"))
	}

	return msg
}
