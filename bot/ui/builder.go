package ui

import (
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

// Builder provides a fluent interface for building Slack UI components.
type Builder struct {
	blocks []slack.Block
	title  *slack.TextBlockObject
	close  *slack.TextBlockObject
	submit *slack.TextBlockObject
}

// New creates a new UI builder.
func New() *Builder {
	return &Builder{
		blocks: []slack.Block{},
	}
}

// SetTitle sets the modal title.
func (b *Builder) SetTitle(title string) *Builder {
	b.title = slack.NewTextBlockObject(slack.PlainTextType, title, false, false)
	return b
}

// SetClose sets the close button text.
func (b *Builder) SetClose(text string) *Builder {
	b.close = slack.NewTextBlockObject(slack.PlainTextType, text, false, false)
	return b
}

// SetSubmit sets the submit button text.
func (b *Builder) SetSubmit(text string) *Builder {
	b.submit = slack.NewTextBlockObject(slack.PlainTextType, text, false, false)
	return b
}

// AddHeader adds a header section to the modal.
func (b *Builder) AddHeader(text string) *Builder {
	headerText := slack.NewTextBlockObject(slack.MarkdownType, text, false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)
	b.blocks = append(b.blocks, headerSection)
	return b
}

// AddSection adds a text section.
func (b *Builder) AddSection(text string) *Builder {
	sectionText := slack.NewTextBlockObject(slack.MarkdownType, text, false, false)
	section := slack.NewSectionBlock(sectionText, nil, nil)
	b.blocks = append(b.blocks, section)
	return b
}

// AddDivider adds a divider block.
func (b *Builder) AddDivider() *Builder {
	b.blocks = append(b.blocks, slack.NewDividerBlock())
	return b
}

// AddContext adds context text.
func (b *Builder) AddContext(text string) *Builder {
	contextText := slack.NewTextBlockObject(slack.MarkdownType, text, false, false)
	contextBlock := slack.NewContextBlock("", contextText)
	b.blocks = append(b.blocks, contextBlock)
	return b
}

// TextInput represents configuration for a text input.
type TextInput struct {
	Label       string
	Placeholder string
	BlockID     string
	ActionID    string
	Initial     string
	Multiline   bool
	Optional    bool
	MinLength   int
	MaxLength   int
}

// AddTextInput adds a text input field.
func (b *Builder) AddTextInput(config TextInput) *Builder {
	label := slack.NewTextBlockObject(slack.PlainTextType, config.Label, false, false)

	var placeholder *slack.TextBlockObject
	if config.Placeholder != "" {
		placeholder = slack.NewTextBlockObject(slack.PlainTextType, config.Placeholder, false, false)
	}

	var element *slack.PlainTextInputBlockElement
	if config.Multiline {
		element = slack.NewPlainTextInputBlockElement(placeholder, config.ActionID)
		element.Multiline = true
	} else {
		element = slack.NewPlainTextInputBlockElement(placeholder, config.ActionID)
	}

	if config.Initial != "" {
		element.InitialValue = config.Initial
	}

	if config.MinLength > 0 {
		element.MinLength = config.MinLength
	}

	if config.MaxLength > 0 {
		element.MaxLength = config.MaxLength
	}

	block := slack.NewInputBlock(config.BlockID, label, nil, element)
	block.Optional = config.Optional

	b.blocks = append(b.blocks, block)
	return b
}

// AddEmailInput adds an email input field.
func (b *Builder) AddEmailInput(blockID, label, placeholder, initial string, optional bool) *Builder {
	return b.AddTextInput(TextInput{
		Label:       label,
		Placeholder: placeholder,
		BlockID:     blockID,
		ActionID:    "email",
		Initial:     initial,
		Optional:    optional,
	})
}

// DatePickerInput represents configuration for a date picker.
type DatePickerInput struct {
	Label       string
	BlockID     string
	ActionID    string
	Initial     *time.Time
	Placeholder string
	Optional    bool
}

// AddDatePicker adds a date picker input.
func (b *Builder) AddDatePicker(config DatePickerInput) *Builder {
	label := slack.NewTextBlockObject(slack.PlainTextType, config.Label, false, false)

	datePicker := slack.NewDatePickerBlockElement(config.ActionID)

	if config.Initial != nil {
		datePicker.InitialDate = config.Initial.Format("2006-01-02")
	}

	var placeholder *slack.TextBlockObject
	if config.Placeholder != "" {
		placeholder = slack.NewTextBlockObject(slack.PlainTextType, config.Placeholder, false, false)
	}

	block := slack.NewInputBlock(config.BlockID, label, placeholder, datePicker)
	block.Optional = config.Optional

	b.blocks = append(b.blocks, block)
	return b
}

// TimePickerInput represents configuration for a time picker.
type TimePickerInput struct {
	Label       string
	BlockID     string
	ActionID    string
	Initial     *time.Time
	Placeholder string
	Optional    bool
}

// AddTimePicker adds a time picker input.
func (b *Builder) AddTimePicker(config TimePickerInput) *Builder {
	label := slack.NewTextBlockObject(slack.PlainTextType, config.Label, false, false)

	timePicker := slack.NewTimePickerBlockElement(config.ActionID)

	if config.Initial != nil {
		timePicker.InitialTime = config.Initial.Format("15:04")
	}

	var placeholder *slack.TextBlockObject
	if config.Placeholder != "" {
		placeholder = slack.NewTextBlockObject(slack.PlainTextType, config.Placeholder, false, false)
	}

	block := slack.NewInputBlock(config.BlockID, label, placeholder, timePicker)
	block.Optional = config.Optional

	b.blocks = append(b.blocks, block)
	return b
}

// RadioOption represents a radio button option.
type RadioOption struct {
	Text  string
	Value string
	Desc  string
}

// AddRadioButtons adds radio button input.
func (b *Builder) AddRadioButtons(blockID, label, actionID string, options []RadioOption, optional bool) *Builder {
	labelText := slack.NewTextBlockObject(slack.PlainTextType, label, false, false)

	var radioOptions []*slack.OptionBlockObject
	for _, opt := range options {
		text := slack.NewTextBlockObject(slack.PlainTextType, opt.Text, false, false)
		option := slack.NewOptionBlockObject(opt.Value, text, nil)
		if opt.Desc != "" {
			desc := slack.NewTextBlockObject(slack.PlainTextType, opt.Desc, false, false)
			option.Description = desc
		}
		radioOptions = append(radioOptions, option)
	}

	radioElement := slack.NewRadioButtonsBlockElement(actionID, radioOptions...)

	block := slack.NewInputBlock(blockID, labelText, nil, radioElement)
	block.Optional = optional

	b.blocks = append(b.blocks, block)
	return b
}

// AddYesNoRadio adds a simple yes/no radio button.
func (b *Builder) AddYesNoRadio(blockID, label, actionID string, optional bool) *Builder {
	return b.AddRadioButtons(blockID, label, actionID, []RadioOption{
		{Text: "Yes", Value: "yes"},
		{Text: "No", Value: "no"},
	}, optional)
}

// SelectOption represents a select menu option.
type SelectOption struct {
	Text  string
	Value string
}

// AddSelectMenu adds a select menu.
func (b *Builder) AddSelectMenu(blockID, label, actionID, placeholder string, options []SelectOption, optional bool) *Builder {
	labelText := slack.NewTextBlockObject(slack.PlainTextType, label, false, false)
	placeholderText := slack.NewTextBlockObject(slack.PlainTextType, placeholder, false, false)

	var selectOptions []*slack.OptionBlockObject
	for _, opt := range options {
		text := slack.NewTextBlockObject(slack.PlainTextType, opt.Text, false, false)
		option := slack.NewOptionBlockObject(opt.Value, text, nil)
		selectOptions = append(selectOptions, option)
	}

	selectElement := slack.NewOptionsSelectBlockElement(slack.OptTypeStatic, placeholderText, actionID, selectOptions...)

	block := slack.NewInputBlock(blockID, labelText, nil, selectElement)
	block.Optional = optional

	b.blocks = append(b.blocks, block)
	return b
}

// AddUserSelect adds a user selection menu.
func (b *Builder) AddUserSelect(blockID, label, actionID, placeholder string, optional bool) *Builder {
	labelText := slack.NewTextBlockObject(slack.PlainTextType, label, false, false)
	placeholderText := slack.NewTextBlockObject(slack.PlainTextType, placeholder, false, false)

	userSelect := slack.NewOptionsSelectBlockElement(
		slack.OptTypeUser,
		placeholderText,
		actionID,
	)

	block := slack.NewInputBlock(blockID, labelText, nil, userSelect)
	block.Optional = optional

	b.blocks = append(b.blocks, block)
	return b
}

// AddChannelSelect adds a channel selection menu.
func (b *Builder) AddChannelSelect(blockID, label, actionID, placeholder string, optional bool) *Builder {
	labelText := slack.NewTextBlockObject(slack.PlainTextType, label, false, false)
	placeholderText := slack.NewTextBlockObject(slack.PlainTextType, placeholder, false, false)

	channelSelect := slack.NewOptionsSelectBlockElement(
		slack.OptTypeConversations,
		placeholderText,
		actionID,
	)

	block := slack.NewInputBlock(blockID, labelText, nil, channelSelect)
	block.Optional = optional

	b.blocks = append(b.blocks, block)
	return b
}

// AddCheckboxes adds checkbox options.
func (b *Builder) AddCheckboxes(blockID, label, actionID string, options []SelectOption, optional bool) *Builder {
	labelText := slack.NewTextBlockObject(slack.PlainTextType, label, false, false)

	var checkOptions []*slack.OptionBlockObject
	for _, opt := range options {
		text := slack.NewTextBlockObject(slack.PlainTextType, opt.Text, false, false)
		option := slack.NewOptionBlockObject(opt.Value, text, nil)
		checkOptions = append(checkOptions, option)
	}

	checkElement := slack.NewCheckboxGroupsBlockElement(actionID, checkOptions...)

	block := slack.NewInputBlock(blockID, labelText, nil, checkElement)
	block.Optional = optional

	b.blocks = append(b.blocks, block)
	return b
}

// ButtonConfig represents button configuration.
type ButtonConfig struct {
	Text     string
	ActionID string
	Value    string
	Style    slack.Style
	Confirm  *ConfirmDialog
	URL      string
}

// ConfirmDialog represents a confirmation dialog.
type ConfirmDialog struct {
	Title   string
	Text    string
	Confirm string
	Deny    string
}

// AddButton adds a button element.
func (b *Builder) AddButton(config ButtonConfig) *Builder {
	buttonText := slack.NewTextBlockObject(slack.PlainTextType, config.Text, false, false)
	button := slack.NewButtonBlockElement(config.ActionID, config.Value, buttonText)

	if config.Style != "" {
		button.Style = config.Style
	}

	if config.URL != "" {
		button.URL = config.URL
	}

	if config.Confirm != nil {
		confirmTitle := slack.NewTextBlockObject(slack.PlainTextType, config.Confirm.Title, false, false)
		confirmText := slack.NewTextBlockObject(slack.PlainTextType, config.Confirm.Text, false, false)
		confirmBtn := slack.NewTextBlockObject(slack.PlainTextType, config.Confirm.Confirm, false, false)
		denyBtn := slack.NewTextBlockObject(slack.PlainTextType, config.Confirm.Deny, false, false)

		button.Confirm = slack.NewConfirmationBlockObject(confirmTitle, confirmText, confirmBtn, denyBtn)
	}

	actionBlock := slack.NewActionBlock("", button)
	b.blocks = append(b.blocks, actionBlock)
	return b
}

// AddActions adds multiple buttons in a single action block.
func (b *Builder) AddActions(buttons ...ButtonConfig) *Builder {
	var elements []slack.BlockElement

	for _, config := range buttons {
		buttonText := slack.NewTextBlockObject(slack.PlainTextType, config.Text, false, false)
		button := slack.NewButtonBlockElement(config.ActionID, config.Value, buttonText)

		if config.Style != "" {
			button.Style = config.Style
		}

		if config.URL != "" {
			button.URL = config.URL
		}

		elements = append(elements, button)
	}

	actionBlock := slack.NewActionBlock("", elements...)
	b.blocks = append(b.blocks, actionBlock)
	return b
}

// BuildBlocks returns the blocks for use in messages.
func (b *Builder) BuildBlocks() slack.Blocks {
	return slack.Blocks{
		BlockSet: b.blocks,
	}
}

// BuildModal creates a modal view request.
func (b *Builder) BuildModal(callbackID string, externalID string, privateMetadata string) slack.ModalViewRequest {
	if b.title == nil {
		b.title = slack.NewTextBlockObject(slack.PlainTextType, "Modal", false, false)
	}

	modal := slack.ModalViewRequest{
		Type:            slack.VTModal,
		Title:           b.title,
		Blocks:          b.BuildBlocks(),
		CallbackID:      callbackID,
		ExternalID:      externalID,
		PrivateMetadata: privateMetadata,
	}

	if b.close != nil {
		modal.Close = b.close
	}

	if b.submit != nil {
		modal.Submit = b.submit
	}

	return modal
}

// BuildMessage creates message options with the blocks.
func (b *Builder) BuildMessage() []slack.MsgOption {
	return []slack.MsgOption{
		slack.MsgOptionBlocks(b.blocks...),
	}
}

// Clear resets the builder.
func (b *Builder) Clear() *Builder {
	b.blocks = []slack.Block{}
	b.title = nil
	b.close = nil
	b.submit = nil
	return b
}

// Clone creates a copy of the builder.
func (b *Builder) Clone() *Builder {
	newBuilder := &Builder{
		blocks: make([]slack.Block, len(b.blocks)),
		title:  b.title,
		close:  b.close,
		submit: b.submit,
	}
	copy(newBuilder.blocks, b.blocks)
	return newBuilder
}

// Helper function to create commonly used modals

// CreateUserUpdateModal creates a standardized user update modal.
func CreateUserUpdateModal(userEmail string, firstAck bool, firstAckTime *time.Time, tzOffset int) slack.ModalViewRequest {
	builder := New().
		SetTitle("Update User").
		SetClose("Cancel").
		SetSubmit("Update").
		AddHeader("Update User Information").
		AddEmailInput("email", "Email", userEmail, userEmail, true)

	// Add first acknowledgment radio
	builder.AddRadioButtons("first_ack", "First Message Acknowledged", "first_ack_option", []RadioOption{
		{Text: "Yes", Value: "yes"},
		{Text: "No", Value: "no"},
	}, true)

	// Add date picker for acknowledgment time
	if firstAckTime != nil {
		builder.AddDatePicker(DatePickerInput{
			Label:    "First Message Acknowledged At",
			BlockID:  "first_ack_at",
			ActionID: "first_ack_date",
			Initial:  firstAckTime,
			Optional: true,
		})
	}

	// Add timezone offset
	tzString := fmt.Sprintf("%d", tzOffset)
	builder.AddTextInput(TextInput{
		Label:       "Timezone Offset",
		Placeholder: tzString,
		BlockID:     "tz_offset",
		ActionID:    "tz_offset",
		Initial:     tzString,
		Optional:    true,
	})

	return builder.BuildModal("user_update_modal", "", userEmail)
}

// CreateReminderModal creates a reminder scheduling modal.
func CreateReminderModal() slack.ModalViewRequest {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)

	return New().
		SetTitle("Set Reminder").
		SetClose("Cancel").
		SetSubmit("Set Reminder").
		AddHeader("Schedule a Reminder").
		AddSection("Choose when you'd like to be reminded to complete this task.").
		AddDivider().
		AddDatePicker(DatePickerInput{
			Label:    "Reminder Date",
			BlockID:  "reminder_date",
			ActionID: "date_select",
			Initial:  &tomorrow,
			Optional: false,
		}).
		AddTimePicker(TimePickerInput{
			Label:    "Reminder Time",
			BlockID:  "reminder_time",
			ActionID: "time_select",
			Initial:  &tomorrow,
			Optional: false,
		}).
		AddContext("You'll receive a reminder at the specified time in your timezone.").
		BuildModal("reminder_modal", "", "")
}
