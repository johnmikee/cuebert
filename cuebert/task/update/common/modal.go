package common

import (
	"github.com/johnmikee/cuebert/cuebert/task/update/db/tracking"
	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/slack-go/slack"
)

const (
	GroupDM = "group_dm"
)

var (
	titleText  = slack.NewTextBlockObject(slack.PlainTextType, "Cuebert", false, false)
	closeText  = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	submitText = slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
	yes        = slack.NewTextBlockObject(slack.PlainTextType, "Yes", false, false)
	no         = slack.NewTextBlockObject(slack.PlainTextType, "No", false, false)
)

func (c *Common) UserUpdateModalSubmit(base *tracking.Info, values map[string]map[string]slack.BlockAction) {
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
	if values[SerialNumber][SerialInput].Value != "" {
		base.SerialNumber = values[SerialInput][SerialNumber].Value
	}
}
