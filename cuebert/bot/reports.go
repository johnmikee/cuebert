package bot

import (
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/visual"
	"github.com/slack-go/slack"
)

func generateReport(fileName, title string) slack.FileUploadParameters {
	return slack.FileUploadParameters{
		Title:    title,
		Filetype: "image/png",
		File:     fileName,
	}
}

func (b *Bot) sendReport(o *visual.PieChartOption, channel ...string) error {
	graph, err := visual.PieChart(o)

	if err != nil {
		return err
	}

	report := generateReport(graph, o.Text)
	report.Channels = channel

	_, err = b.bot.SlackClient().UploadFile(report)

	return err
}

// SendDailyReport sends a daily report to the configured channel
// with a pie chart of the OS breakdown of all devices in the database.
func (b *Bot) SendDailyAdminReport() error {
	r, err := b.BuildOSReport()
	if err != nil {
		b.log.Debug().AnErr("building report", err).
			Send()
		return err
	}

	err = b.sendReport(r, b.cfg.slackAlertChannel)

	if err != nil {
		b.log.Debug().AnErr("sending report", err).
			Send()
	}

	return err
}

// Report is used to determine which report to build.
type Report string

const (
	First             Report = "first"
	Manager           Report = "manager"
	ReminderRequested Report = "reminderRequested"
)

// buildSentReport builds a report of the number of users that have been sent a message.
func (b *Bot) BuildSentReport(which Report) (*visual.PieChartOption, error) {
	br, err := b.tables.GetBotTableInfo()
	if err != nil {
		b.log.Debug().AnErr("getting br", err).
			Send()
		return nil, err
	}

	v := &visual.PieChartOption{}

	var sent, notSent int

	switch which {
	case First:
		sent, notSent = countSentStatus(br, First)
		v.Query = "FirstMessageSent"
		v.Text = "First Message Sent"
	case Manager:
		sent, notSent = countSentStatus(br, Manager)
		v.Query = "ManagerMessageSent"
		v.Text = "Manager Message Sent"
	case ReminderRequested:
		sent, notSent = countSentStatus(br, ReminderRequested)
		v.Query = "RequestedReminder"
		v.Text = "Requested Reminder"
	}

	v.ValueList = append(v.ValueList, float64(sent), float64(notSent))
	v.XAxis = append(v.XAxis, "Sent", "Not Sent")

	return v, nil
}

// buildOSReport builds a report of the number of devices by OS.
func (b *Bot) BuildOSReport() (*visual.PieChartOption, error) {
	br, err := b.tables.GetAllDevices()
	if err != nil {
		return nil, err
	}

	v := &visual.PieChartOption{}
	rep := make(map[string]int)

	for i := range br {
		if rep[br[i].OSVersion] == 0 {
			rep[br[i].OSVersion] = 1
		} else {
			rep[br[i].OSVersion]++
		}
	}

	for k, val := range rep {
		f := float64(val)
		v.ValueList = append(v.ValueList, f)
		v.XAxis = append(v.XAxis, k)
	}

	v.Query = "OSVersion"
	v.Text = "OS Version"

	return v, nil
}

func countSentStatus(br bot.BR, which Report) (sent, notSent int) {
	for i := range br {
		switch which {
		case First:
			if br[i].FirstMessageSent {
				sent++
			} else {
				notSent++
			}
		case Manager:
			if br[i].ManagerMessageSent {
				sent++
			} else {
				notSent++
			}
		case ReminderRequested:
			if !br[i].DelayAt.IsZero() {
				sent++
			} else {
				notSent++
			}
		}
	}
	return sent, notSent
}
