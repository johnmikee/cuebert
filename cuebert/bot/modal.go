package bot

import (
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

type modalGateway struct {
	text       string
	fallback   string
	callbackID string
	yesName    string
	yesText    string
	yesValue   string
	yesStyle   string
	noName     string
	noText     string
	noValue    string
	noStyle    string
	channel    string
	msg        string
}

var (
	titleText  = slack.NewTextBlockObject(slack.PlainTextType, "Cuebert", false, false)
	closeText  = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	submitText = slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
	yes        = slack.NewTextBlockObject(slack.PlainTextType, "Yes", false, false)
	no         = slack.NewTextBlockObject(slack.PlainTextType, "No", false, false)
)

// modalGateway is a helper function to send a request to invoke a modal with two buttons.
func (b *Bot) modalGateway(m *modalGateway) {
	attachment := slack.Attachment{
		Text:       m.text,
		Fallback:   m.fallback,
		CallbackID: m.callbackID,
		Color:      "#3AA3E3",
		Actions: []slack.AttachmentAction{
			{
				Name:  m.yesName,
				Text:  m.yesText,
				Type:  "button",
				Value: m.yesValue,
				Style: m.yesStyle,
			},
			{
				Name:  m.noName,
				Text:  m.noText,
				Type:  "button",
				Value: m.noValue,
				Style: m.noStyle,
			},
		},
	}

	b.sendAttachment(attachment, m.channel, m.msg)
}

// interactiveHelper is the handler for all modal requests
func (b *Bot) interactiveHelper(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0]

	switch action.Name {
	case Accept:
		b.ack(ctx)
	case UpdateUserYes:
		b.userSelector(ctx.Callback().TriggerID, ctx.Callback().Channel.ID)
	case UpdateUserSubmit:
		// maybe this is fixed now?
		// it was b.UserBRUpdate(s, event, callback, callback.Channel.ID)
		b.method.UserUpdate(ctx)
	case UpdateConfigYes:
		b.loadProgram(ctx.Callback().TriggerID, ctx.Callback().CallbackID)
	}
	_, _, _ = b.Client().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
}

// modalSubmit is the handler for all modal submissions
func (b *Bot) modalSubmit(ctx *slacker.InteractionContext) {
	switch ctx.Callback().View.CallbackID {
	case AdminExclusionModal:
		b.exclusionSubmit(ctx)
	case ExclusionModal:
		b.userExclusionSubmit(ctx)
	case ReminderPicker:
		b.reminderSubmit(ctx)
	case UpdateUserSelector:
		// maybe this is fixed now?
		// it was  b.userBRUpdate(s, event, callback, callback.TriggerID)
		b.method.UserUpdate(ctx)
	case UserUpdateModal:
		b.userUpdateSubmit(ctx)
	case Start, Reload:
		b.loadInput(ctx, ctx.Callback().View.CallbackID)
	default:
		b.log.Debug().Str("callback", ctx.Callback().View.CallbackID).Msg("not something we are handling.")
		return
	}
}

// createOptionBlockObjects - utility function for generating option block objects
func createOptionBlockObjects(options []string) []*slack.OptionBlockObject {
	optionBlockObjects := make([]*slack.OptionBlockObject, 0, len(options))

	for _, o := range options {
		optionText := slack.NewTextBlockObject(slack.PlainTextType, o, false, false)
		optionBlockObjects = append(optionBlockObjects, slack.NewOptionBlockObject(o, optionText, nil))
	}

	return optionBlockObjects
}
