package bot

import (
	"fmt"

	"github.com/shomali11/slacker/v2"
)

type Time struct{}

func (Time) Minute(interval ...int) string {
	if len(interval) > 0 {
		return fmt.Sprintf("*/%d * * * *", interval[0])
	}
	return "*/1 * * * *"
}

func (Time) Hour(interval ...int) string {
	if len(interval) > 0 {
		if interval[0] <= 23 {
			return fmt.Sprintf("0 %d * * *", interval[0])
		}
	}
	return "0 * * * *"
}

var Every Time

func (b *Bot) AddJob(j *slacker.JobDefinition) {
	b.bot.AddJob(j)
}
