package shared

import (
	"github.com/slack-go/slack"
)

func datePickerBlock(until, callback string) *slack.InputBlock {
	datePickerSection := slack.NewTextBlockObject(slack.PlainTextType, "Until", false, false)
	datePickerHint := slack.NewTextBlockObject(slack.PlainTextType, until, false, false)
	datePickerInput := slack.NewDatePickerBlockElement(DatePicker)
	datePickerBlock := slack.NewInputBlock(callback, datePickerSection, datePickerHint, datePickerInput)

	return datePickerBlock
}

func headerBlock(title string) *slack.SectionBlock {
	headerText := slack.NewTextBlockObject(slack.MarkdownType, title, false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	return headerSection
}

func reasonBlock(reason, hint, callback string) *slack.InputBlock {
	reasonSection := slack.NewTextBlockObject(slack.PlainTextType, reason, false, false)
	reasonHint := slack.NewTextBlockObject(slack.PlainTextType, hint, false, false)
	reasonInput := slack.NewPlainTextInputBlockElement(reasonHint, ExclusionInput)
	reasonBlock := slack.NewInputBlock(callback, reasonSection, nil, reasonInput)

	return reasonBlock
}
