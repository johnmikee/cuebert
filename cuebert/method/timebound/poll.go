package timebound

import (
	"sync"
	"time"

	"github.com/johnmikee/cuebert/cuebert/bot"
	br "github.com/johnmikee/cuebert/db/bot"
	"github.com/shomali11/slacker/v2"
)

var (
	once sync.Once
)

// register jobs on our intervals once. these will take a slice of users to remind on
// the interval. the poll checks that the groups are still valid based on user interval preference
// and moves them as needed.

func (t *TimeBound) Poll(time.Time) {
	// make sure the jobs are only started once
	once.Do(func() {
		t.registerCronJobs()
	})

	t.pollGroups()
}

func (t *TimeBound) registerCronJobs() {
	t.base()
	t.thirty()
	t.hourly()
	t.twoHour()
	t.fourHour()
}

func (t *TimeBound) base() {
	t.bot.AddJob(&slacker.JobDefinition{
		CronExpression: bot.Every.Minute(t.cfg.defaultReminderInterval),
		Name:           "30M",
		Description:    "Sends reminders every 30 minutes",
		Handler: func(ctx *slacker.JobContext) {
			t.poke(ctx, t.cfg.defaultReminderInterval)
		},
		HideHelp: true,
	})
}

func (t *TimeBound) thirty() {
	t.bot.AddJob(&slacker.JobDefinition{
		CronExpression: bot.Every.Minute(30),
		Name:           "30M",
		Description:    "Sends reminders every 30 minutes",
		Handler: func(ctx *slacker.JobContext) {
			t.poke(ctx, 30)
		},
		HideHelp: true,
	})
}

func (t *TimeBound) hourly() {
	t.bot.AddJob(&slacker.JobDefinition{
		CronExpression: bot.Every.Hour(),
		Name:           "1H",
		Description:    "Sends reminders every hour",
		Handler: func(ctx *slacker.JobContext) {
			t.poke(ctx, 60)
		},
		HideHelp: true,
	})
}

func (t *TimeBound) twoHour() {
	t.bot.AddJob(&slacker.JobDefinition{
		CronExpression: bot.Every.Hour(2),
		Name:           "2H",
		Description:    "Sends reminders every two hours",
		Handler: func(ctx *slacker.JobContext) {
			t.poke(ctx, 120)
		},
		HideHelp: true,
	})
}

func (t *TimeBound) fourHour() {
	t.bot.AddJob(&slacker.JobDefinition{
		CronExpression: bot.Every.Hour(4),
		Name:           "4H",
		Description:    "Sends reminders every four hours",
		Handler: func(ctx *slacker.JobContext) {
			t.poke(ctx, 240)
		},
		HideHelp: true,
	})
}

// poke them with a reminder
func (t *TimeBound) poke(ctx *slacker.JobContext, ri int) {
	var devices br.BR
	switch ri {
	case 30:
		devices = t.intervals.thirty
	case 60:
		devices = t.intervals.hour
	case 120:
		devices = t.intervals.two
	case 240:
		devices = t.intervals.four
	default:
		devices = t.intervals.base
	}

	for i := range devices {
		rw, err := t.tables.GetReminderWaiting(devices[i].SerialNumber)
		if err != nil {
			t.log.Error().
				Str("user", devices[i].FullName).
				Str("serial", devices[i].SerialNumber).
				AnErr("could not get reminder waiting", err).
				Msg("skipping")
			continue
		}

		if rw {
			continue
		}

		go t.waitSend(&devices[i])
	}

}

func (t *TimeBound) pollGroups() {
	devices, err := t.tables.GetBotTableInfo()
	if err != nil {
		return
	}

	for i := range devices {
		ri, err := t.tables.GetReminderInterval(devices[i].SerialNumber)
		if err != nil {
			t.log.Error().
				Str("user", devices[i].FullName).
				Str("serial", devices[i].SerialNumber).
				AnErr("could not get reminder interval", err).
				Msg("skipping")
			continue
		}

		switch ri {
		case 30:
			if !contains(t.intervals.thirty, devices[i].SerialNumber) {
				t.intervals.thirty = append(t.intervals.thirty, devices[i])
				removeFromOthers(
					devices[i].SerialNumber,
					&t.intervals.thirty,
					&t.intervals.hour,
					&t.intervals.two,
					&t.intervals.four,
					&t.intervals.base,
				)
			}
		case 60:
			if !contains(t.intervals.hour, devices[i].SerialNumber) {
				t.intervals.hour = append(t.intervals.hour, devices[i])
				removeFromOthers(
					devices[i].SerialNumber,
					&t.intervals.hour,
					&t.intervals.thirty,
					&t.intervals.two,
					&t.intervals.four,
					&t.intervals.base,
				)
			}
		case 120:
			if !contains(t.intervals.two, devices[i].SerialNumber) {
				t.intervals.two = append(t.intervals.two, devices[i])
				removeFromOthers(
					devices[i].SerialNumber,
					&t.intervals.two,
					&t.intervals.thirty,
					&t.intervals.hour,
					&t.intervals.four,
					&t.intervals.base,
				)
			}
		case 240:
			if !contains(t.intervals.four, devices[i].SerialNumber) {
				t.intervals.four = append(t.intervals.four, devices[i])
				removeFromOthers(
					devices[i].SerialNumber,
					&t.intervals.four,
					&t.intervals.thirty,
					&t.intervals.hour,
					&t.intervals.two,
					&t.intervals.base,
				)
			}
		default:
			if !contains(t.intervals.base, devices[i].SerialNumber) {
				t.intervals.base = append(t.intervals.base, devices[i])
				removeFromOthers(
					devices[i].SerialNumber,
					&t.intervals.base,
					&t.intervals.thirty,
					&t.intervals.hour,
					&t.intervals.two,
					&t.intervals.four,
				)
			}
		}
	}
}

func contains(info br.BR, target string) bool {
	for _, i := range info {
		if i.SerialNumber == target {
			return true
		}
	}
	return false
}

func removeFromOthers(target string, targetArray *br.BR, arrays ...*br.BR) {
	for _, arr := range arrays {
		if arr != targetArray {
			remove(target, arr)
		}
	}
}

func remove(target string, array *br.BR) {
	for i, info := range *array {
		if info.SerialNumber == target {
			(*array) = append((*array)[:i], (*array)[i+1:]...)
			break
		}
	}
}
