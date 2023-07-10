package manager

import (
	"time"

	"github.com/johnmikee/cuebert/pkg/helpers"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

const (
	Email                    = "email"
	FirstMessageACK          = "first_message_ack"
	FirstMessageACKAt        = "first_message_ack_at"
	FirstMessageACKAtPicker  = "first_message_ack_at_picker"
	FirstMessageACKOption    = "first_message_ack_option"
	FirstMessageSentAt       = "first_message_sent_at"
	FirstMessageSentAtPicker = "first_message_sent_at_picker"
	FirstMessageSentOption   = "first_message_sent_option"
	FirstSent                = "first_message_sent"
	Reload                   = "reload_settings_modal"
	RemindMe                 = "remind_me"
	SerialNumber             = "serial_number"
	Start                    = "start_settings_modal"
	TruthyNo                 = "false"
	TruthyYes                = "true"
	TZOffset                 = "tz_offset"
)

var (
	titleText  = slack.NewTextBlockObject(slack.PlainTextType, "Cuebert", false, false)
	closeText  = slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false)
	submitText = slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
	yes        = slack.NewTextBlockObject(slack.PlainTextType, "Yes", false, false)
	no         = slack.NewTextBlockObject(slack.PlainTextType, "No", false, false)
)

func (m *Manager) UserUpdate(ctx *slacker.InteractionContext) {
	values := ctx.Callback().View.State.Values
	userID := values[UpdateUser][UsersSelect].SelectedUser

	user, err := m.tables.UserBySlackID(userID)
	if err != nil {
		m.log.Debug().AnErr("getting user", err).Send()
		return
	}
	if user.Empty() {
		m.log.Debug().Msg("user not found")
		return
	}

	userInfo := user[0]

	headerText := slack.NewTextBlockObject(slack.MarkdownType, "Update User", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// now we load the values from the user and set them as the default values for the input blocks
	email := slack.NewTextBlockObject(slack.PlainTextType, "Email", false, false)
	emailPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.UserEmail, false, false)
	emailInput := slack.NewPlainTextInputBlockElement(emailPlaceHolder, "email")
	emailBlock := slack.NewInputBlock(Email, email, nil, emailInput)
	emailBlock.Optional = true

	managerSID := slack.NewTextBlockObject(slack.PlainTextType, "Manager Slack ID", false, false)
	managerOptBlock := slack.NewOptionsSelectBlockElement(slack.OptTypeUser, nil, UsersSelect)
	managerSIDBlock := slack.NewInputBlock(ManagerSlackID, managerSID, nil, managerOptBlock)
	managerSIDBlock.Optional = true

	firstAck := slack.NewTextBlockObject(slack.PlainTextType, "First Message Acknowledged", false, false)
	firstAckSentYes := slack.NewOptionBlockObject(TruthyYes, yes, nil)
	firstAckSentNo := slack.NewOptionBlockObject(TruthyNo, no, nil)
	firstAckOptions := slack.NewRadioButtonsBlockElement(FirstMessageACKOption, firstAckSentYes, firstAckSentNo)
	firstAckBlock := slack.NewInputBlock(FirstMessageACK, firstAck, nil, firstAckOptions)
	firstAckBlock.Optional = true

	firstAckAt := slack.NewTextBlockObject(slack.PlainTextType, "First Message Acknowledged At", false, false)
	firstAckAtPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.FirstACKTime.Format(time.RFC3339), false, false)
	firstAckAtDateBlock := slack.NewDatePickerBlockElement(FirstMessageACKAt)
	firstAckAtInput := slack.NewInputBlock(FirstMessageACKAtPicker, firstAckAt, firstAckAtPlaceHolder, firstAckAtDateBlock)
	firstAckAtInput.Optional = true

	firstSent := slack.NewTextBlockObject(slack.PlainTextType, "First Message Sent", false, false)
	firstSentYes := slack.NewOptionBlockObject(TruthyYes, yes, nil)
	firstSentNo := slack.NewOptionBlockObject(TruthyNo, no, nil)
	firstSentOptions := slack.NewRadioButtonsBlockElement(FirstMessageSentOption, firstSentYes, firstSentNo)
	firstSentBlock := slack.NewInputBlock(FirstMessageSentOption, firstSent, nil, firstSentOptions)
	firstSentBlock.Optional = true

	firstSentAt := slack.NewTextBlockObject(slack.PlainTextType, "First Message Sent At", false, false)
	firstSentAtPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.FirstMessageSentAt.Format(time.RFC3339), false, false)
	firstSentAtDateBlock := slack.NewDatePickerBlockElement(FirstMessageSentAt)
	firstSentAtInput := slack.NewInputBlock(FirstMessageSentAtPicker, firstSentAt, firstSentAtPlaceHolder, firstSentAtDateBlock)
	firstSentAtInput.Optional = true

	managerMessageSent := slack.NewTextBlockObject(slack.PlainTextType, "Manager Message Sent", false, false)
	managerMessageSentYes := slack.NewOptionBlockObject(TruthyYes, yes, nil)
	managerMessageSentNo := slack.NewOptionBlockObject(TruthyNo, no, nil)
	managerMessageSentOptions := slack.NewRadioButtonsBlockElement(ManagerMessageSentOption, managerMessageSentYes, managerMessageSentNo)
	managerMessageSentBlock := slack.NewInputBlock(ManagerMessageSentOption, managerMessageSent, nil, managerMessageSentOptions)
	managerMessageSentBlock.Optional = true

	managerMessageSentAt := slack.NewTextBlockObject(slack.PlainTextType, "Manager Message Sent At", false, false)
	managerMessageSentAtPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.ManagerMessageSentAt.Format(time.RFC3339), false, false)
	managerMessageSentAtDateBlock := slack.NewDatePickerBlockElement(ManagerMessageSentAt)
	managerMessageSentAtInput := slack.NewInputBlock(ManagerMessageSentAtPicker, managerMessageSentAt, managerMessageSentAtPlaceHolder, managerMessageSentAtDateBlock)
	managerMessageSentAtInput.Optional = true

	serial := slack.NewTextBlockObject(slack.PlainTextType, "Serial Number", false, false)
	serialPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, userInfo.SerialNumber, false, false)
	serialInput := slack.NewPlainTextInputBlockElement(serialPlaceHolder, SerialNumber)
	serialBlock := slack.NewInputBlock(SerialNumber, serial, nil, serialInput)
	serialBlock.Optional = true

	tzOffset := slack.NewTextBlockObject(slack.PlainTextType, "Timezone Offset", false, false)
	tzOffsetPlaceHolder := slack.NewTextBlockObject(slack.PlainTextType, helpers.TZStringer(userInfo.TZOffset), false, false)
	tzOffsetInput := slack.NewPlainTextInputBlockElement(tzOffsetPlaceHolder, "tz_offset")
	tzOffsetBlock := slack.NewInputBlock("tz_offset", tzOffset, nil, tzOffsetInput)
	tzOffsetBlock.Optional = true

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			emailBlock,
			managerSIDBlock,
			firstAckBlock,
			firstAckAtInput,
			firstSentBlock,
			firstSentAtInput,
			managerMessageSentBlock,
			managerMessageSentAtInput,
			serialBlock,
			tzOffsetBlock,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           titleText,
		Close:           closeText,
		Submit:          submitText,
		Blocks:          blocks,
		CallbackID:      UserUpdateModal,
		PrivateMetadata: userInfo.SlackID,
		ExternalID:      ctx.Callback().View.ExternalID,
	}

	m.log.Info().Interface("modal_request", modalRequest).Msg("opening modal")
	vr, err := m.sc.OpenView(ctx.Callback().TriggerID, modalRequest)
	if err != nil {
		m.log.Error().Err(err).Msg("error opening modal")
	}

	m.log.Trace().Interface("view_response", vr).Msg("modal opened")
}
