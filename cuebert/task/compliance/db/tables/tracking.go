package tables

import (
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/compliance/db/tracking"
)

// SetReminder sets the reminder time for the given slack id.
func (b *Config) SetReminder(id string, t time.Time, ds, ts string) error {
	_, err := b.Tracking().Update().
		WithOptions(
			tracking.Info{
				DelayAt:   t,
				DelayDate: ds,
				DelayTime: ts,
			},
		).
		Parse(tracking.UserSlackID, id).Send()

	return err
}
