package manager

import (
	"github.com/johnmikee/cuebert/db/bot"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/slack-go/slack"
)

// updateManagerVals updates the manager values in the db.
func (m *Manager) updateManagerVals(update bot.BR) error {
	m.log.Debug().Msg("updating manager values")

	for i := range update {
		m.log.Debug().
			Str("user", update[i].UserEmail).
			Str("slack", update[i].SlackID).
			Str("manager", update[i].ManagerSlackID).
			Msg("updating manager value")
		err := m.tables.AddManagerID(update[i].SlackID, update[i].UserEmail, update[i].ManagerSlackID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) UserUpdateModalSubmit(base *bot.Info, values map[string]map[string]slack.BlockAction) {
	if values[Email][Email].Value != "" {
		base.UserEmail = values[Email][Email].Value
	}
	if values[FirstMessageACK][FirstMessageACKOption].SelectedOption.Value != "" {
		base.FirstACK = helpers.StringBooler(values[FirstMessageACK][FirstMessageACKOption].SelectedOption.Value)
	}
	if values[FirstMessageACKAtPicker][FirstMessageACKAt].SelectedOption.Value != "" {
		ts, _ := helpers.StringToTime(values[FirstMessageACKAtPicker][FirstMessageACKAt].SelectedDate)
		base.FirstACKTime = ts
	}
	if values[FirstMessageSentOption][FirstMessageSentOption].SelectedOption.Value != "" {
		base.FirstMessageSent = helpers.StringBooler(values[FirstMessageSentAt][FirstMessageSentAt].SelectedOption.Value)
	}
	if values[ManagerMessageSentAtPicker][ManagerMessageSentAt].SelectedDate != "" {
		ts, _ := helpers.StringToTime(values[ManagerMessageSentAtPicker][ManagerMessageSentAt].SelectedDate)
		base.ManagerMessageSentAt = ts
	}
	if values[ManagerMessageSentOption][ManagerMessageSentOption].SelectedOption.Value != "" {
		base.ManagerMessageSent = helpers.StringBooler(values[ManagerMessageSentOption][ManagerMessageSentOption].SelectedOption.Value)
	}
	if values[ManagerSlackID][UsersSelect].SelectedOption.Value != "" {
		base.ManagerSlackID = values[ManagerSlackID][ManagerSlackID].SelectedOption.Value
	}
	if values[SerialNumber][SerialNumber].Value != "" {
		base.SerialNumber = values[SerialNumber][SerialNumber].Value
	}
}
