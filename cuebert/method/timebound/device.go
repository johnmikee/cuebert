package timebound

import (
	"github.com/johnmikee/cuebert/db/bot"
)

// DeviceDiff removes entries from the intervals if they have updated since the last check.
func (t *TimeBound) DeviceDiff(devices []string) {
	for _, r := range devices {
		removeFromOthers(
			r,
			&bot.BR{},
			&t.intervals.thirty,
			&t.intervals.hour,
			&t.intervals.two,
			&t.intervals.four,
			&t.intervals.base,
		)
	}
}
