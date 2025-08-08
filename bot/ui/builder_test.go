package ui

import (
	"testing"
	"time"

	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuilder_SetTitle(t *testing.T) {
	builder := New()
	builder.SetTitle("Test Title")
	
	assert.NotNil(t, builder.title)
	assert.Equal(t, "Test Title", builder.title.Text)
	assert.Equal(t, slack.PlainTextType, builder.title.Type)
}

func TestBuilder_SetClose(t *testing.T) {
	builder := New()
	builder.SetClose("Cancel")
	
	assert.NotNil(t, builder.close)
	assert.Equal(t, "Cancel", builder.close.Text)
}

func TestBuilder_SetSubmit(t *testing.T) {
	builder := New()
	builder.SetSubmit("Save")
	
	assert.NotNil(t, builder.submit)
	assert.Equal(t, "Save", builder.submit.Text)
}

func TestBuilder_AddHeader(t *testing.T) {
	builder := New()
	builder.AddHeader("Test Header")
	
	require.Len(t, builder.blocks, 1)
	
	section, ok := builder.blocks[0].(*slack.SectionBlock)
	require.True(t, ok)
	assert.Equal(t, "Test Header", section.Text.Text)
	assert.Equal(t, slack.MarkdownType, section.Text.Type)
}

func TestBuilder_AddSection(t *testing.T) {
	builder := New()
	builder.AddSection("This is a test section")
	
	require.Len(t, builder.blocks, 1)
	
	section, ok := builder.blocks[0].(*slack.SectionBlock)
	require.True(t, ok)
	assert.Equal(t, "This is a test section", section.Text.Text)
}

func TestBuilder_AddDivider(t *testing.T) {
	builder := New()
	builder.AddDivider()
	
	require.Len(t, builder.blocks, 1)
	
	_, ok := builder.blocks[0].(*slack.DividerBlock)
	assert.True(t, ok)
}

func TestBuilder_AddContext(t *testing.T) {
	builder := New()
	builder.AddContext("Context text")
	
	require.Len(t, builder.blocks, 1)
	
	context, ok := builder.blocks[0].(*slack.ContextBlock)
	require.True(t, ok)
	require.Len(t, context.ContextElements.Elements, 1)
}

func TestBuilder_AddTextInput(t *testing.T) {
	tests := []struct {
		name   string
		config TextInput
		verify func(t *testing.T, block *slack.InputBlock)
	}{
		{
			name: "basic text input",
			config: TextInput{
				Label:    "Name",
				BlockID:  "name_block",
				ActionID: "name_action",
			},
			verify: func(t *testing.T, block *slack.InputBlock) {
				assert.Equal(t, "Name", block.Label.Text)
				assert.Equal(t, "name_block", block.BlockID)
				
				element, ok := block.Element.(*slack.PlainTextInputBlockElement)
				require.True(t, ok)
				assert.Equal(t, "name_action", element.ActionID)
				assert.False(t, element.Multiline)
			},
		},
		{
			name: "multiline text input with placeholder",
			config: TextInput{
				Label:       "Description",
				Placeholder: "Enter description",
				BlockID:     "desc_block",
				ActionID:    "desc_action",
				Multiline:   true,
				Optional:    true,
			},
			verify: func(t *testing.T, block *slack.InputBlock) {
				assert.Equal(t, "Description", block.Label.Text)
				assert.True(t, *block.Optional)
				
				element, ok := block.Element.(*slack.PlainTextInputBlockElement)
				require.True(t, ok)
				assert.True(t, element.Multiline)
				assert.Equal(t, "Enter description", element.Placeholder.Text)
			},
		},
		{
			name: "text input with initial value and length constraints",
			config: TextInput{
				Label:     "Username",
				BlockID:   "username_block",
				ActionID:  "username_action",
				Initial:   "john_doe",
				MinLength: 3,
				MaxLength: 20,
			},
			verify: func(t *testing.T, block *slack.InputBlock) {
				element, ok := block.Element.(*slack.PlainTextInputBlockElement)
				require.True(t, ok)
				assert.Equal(t, "john_doe", element.InitialValue)
				assert.Equal(t, 3, element.MinLength)
				assert.Equal(t, 20, element.MaxLength)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := New()
			builder.AddTextInput(tt.config)
			
			require.Len(t, builder.blocks, 1)
			
			inputBlock, ok := builder.blocks[0].(*slack.InputBlock)
			require.True(t, ok)
			
			tt.verify(t, inputBlock)
		})
	}
}

func TestBuilder_AddEmailInput(t *testing.T) {
	builder := New()
	builder.AddEmailInput("email_block", "Email", "test@example.com", "user@domain.com", false)
	
	require.Len(t, builder.blocks, 1)
	
	inputBlock, ok := builder.blocks[0].(*slack.InputBlock)
	require.True(t, ok)
	assert.Equal(t, "Email", inputBlock.Label.Text)
	assert.Equal(t, "email_block", inputBlock.BlockID)
	
	element, ok := inputBlock.Element.(*slack.EmailTextInputBlockElement)
	require.True(t, ok)
	assert.Equal(t, "test@example.com", element.InitialValue)
	assert.Equal(t, "user@domain.com", element.Placeholder.Text)
}

func TestBuilder_AddDatePicker(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name   string
		config DatePickerInput
		verify func(t *testing.T, element *slack.DatePickerBlockElement)
	}{
		{
			name: "date picker without initial date",
			config: DatePickerInput{
				Label:    "Start Date",
				BlockID:  "start_date",
				ActionID: "date_select",
				Optional: false,
			},
			verify: func(t *testing.T, element *slack.DatePickerBlockElement) {
				assert.Equal(t, "date_select", element.ActionID)
				assert.Equal(t, "", element.InitialDate)
			},
		},
		{
			name: "date picker with initial date",
			config: DatePickerInput{
				Label:    "End Date",
				BlockID:  "end_date",
				ActionID: "end_date_select",
				Initial:  &now,
				Optional: true,
			},
			verify: func(t *testing.T, element *slack.DatePickerBlockElement) {
				assert.Equal(t, "end_date_select", element.ActionID)
				assert.Equal(t, now.Format("2006-01-02"), element.InitialDate)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := New()
			builder.AddDatePicker(tt.config)
			
			require.Len(t, builder.blocks, 1)
			
			inputBlock, ok := builder.blocks[0].(*slack.InputBlock)
			require.True(t, ok)
			assert.Equal(t, tt.config.Label, inputBlock.Label.Text)
			
			element, ok := inputBlock.Element.(*slack.DatePickerBlockElement)
			require.True(t, ok)
			
			tt.verify(t, element)
		})
	}
}

func TestBuilder_AddRadioButtons(t *testing.T) {
	builder := New()
	options := []RadioOption{
		{Text: "Option 1", Value: "opt1", Desc: "First option"},
		{Text: "Option 2", Value: "opt2", Desc: "Second option"},
		{Text: "Option 3", Value: "opt3"},
	}
	
	builder.AddRadioButtons("radio_block", "Choose one", "radio_action", options, true)
	
	require.Len(t, builder.blocks, 1)
	
	inputBlock, ok := builder.blocks[0].(*slack.InputBlock)
	require.True(t, ok)
	assert.Equal(t, "Choose one", inputBlock.Label.Text)
	assert.True(t, *inputBlock.Optional)
	
	element, ok := inputBlock.Element.(*slack.RadioButtonsBlockElement)
	require.True(t, ok)
	assert.Equal(t, "radio_action", element.ActionID)
	require.Len(t, element.Options, 3)
	
	// Verify options
	assert.Equal(t, "Option 1", element.Options[0].Text.Text)
	assert.Equal(t, "opt1", element.Options[0].Value)
	assert.Equal(t, "First option", element.Options[0].Description.Text)
	
	// Option without description
	assert.Nil(t, element.Options[2].Description)
}

func TestBuilder_AddCheckboxes(t *testing.T) {
	builder := New()
	options := []SelectOption{
		{Text: "Subscribe to newsletter", Value: "newsletter"},
		{Text: "Receive notifications", Value: "notifications"},
	}
	
	builder.AddCheckboxes("checkbox_block", "Preferences", "checkbox_action", options, false)
	
	require.Len(t, builder.blocks, 1)
	
	inputBlock, ok := builder.blocks[0].(*slack.InputBlock)
	require.True(t, ok)
	
	element, ok := inputBlock.Element.(*slack.CheckboxGroupsBlockElement)
	require.True(t, ok)
	assert.Equal(t, "checkbox_action", element.ActionID)
	require.Len(t, element.Options, 2)
}

func TestBuilder_AddSelectMenu(t *testing.T) {
	builder := New()
	options := []SelectOption{
		{Text: "Small", Value: "S"},
		{Text: "Medium", Value: "M"},
		{Text: "Large", Value: "L"},
	}
	
	builder.AddSelectMenu("size_block", "Size", "size_select", "Choose size", options, false)
	
	require.Len(t, builder.blocks, 1)
	
	inputBlock, ok := builder.blocks[0].(*slack.InputBlock)
	require.True(t, ok)
	
	element, ok := inputBlock.Element.(*slack.SelectBlockElement)
	require.True(t, ok)
	assert.Equal(t, "size_select", element.ActionID)
	assert.Equal(t, "Choose size", element.Placeholder.Text)
	require.Len(t, element.Options, 3)
}

func TestBuilder_AddUserSelect(t *testing.T) {
	builder := New()
	builder.AddUserSelect("user_block", "Assign to", "user_select", nil, true)
	
	require.Len(t, builder.blocks, 1)
	
	inputBlock, ok := builder.blocks[0].(*slack.InputBlock)
	require.True(t, ok)
	assert.True(t, *inputBlock.Optional)
	
	element, ok := inputBlock.Element.(*slack.SelectBlockElement)
	require.True(t, ok)
	assert.Equal(t, "users_select", element.Type)
}

func TestBuilder_AddYesNoRadio(t *testing.T) {
	builder := New()
	builder.AddYesNoRadio("confirm_block", "Confirm action?", "confirm_radio", false)
	
	require.Len(t, builder.blocks, 1)
	
	inputBlock, ok := builder.blocks[0].(*slack.InputBlock)
	require.True(t, ok)
	
	element, ok := inputBlock.Element.(*slack.RadioButtonsBlockElement)
	require.True(t, ok)
	require.Len(t, element.Options, 2)
	assert.Equal(t, "Yes", element.Options[0].Text.Text)
	assert.Equal(t, "yes", element.Options[0].Value)
	assert.Equal(t, "No", element.Options[1].Text.Text)
	assert.Equal(t, "no", element.Options[1].Value)
}

func TestBuilder_BuildModal(t *testing.T) {
	builder := New()
	builder.
		SetTitle("Test Modal").
		SetClose("Cancel").
		SetSubmit("Save").
		AddHeader("User Information").
		AddTextInput(TextInput{
			Label:    "Name",
			BlockID:  "name",
			ActionID: "name_input",
		})
	
	modal := builder.BuildModal("test_callback", "external123", "private456")
	
	assert.Equal(t, slack.VTModal, modal.Type)
	assert.Equal(t, "Test Modal", modal.Title.Text)
	assert.Equal(t, "Cancel", modal.Close.Text)
	assert.Equal(t, "Save", modal.Submit.Text)
	assert.Equal(t, "test_callback", modal.CallbackID)
	assert.Equal(t, "external123", modal.ExternalID)
	assert.Equal(t, "private456", modal.PrivateMetadata)
	assert.Len(t, modal.Blocks.BlockSet, 2)
}

func TestBuilder_BuildMessage(t *testing.T) {
	builder := New()
	builder.
		AddSection("Hello, this is a message!").
		AddDivider().
		AddContext("Sent from Cuebert")
	
	message := builder.BuildMessage()
	
	assert.Len(t, message.Blocks.BlockSet, 3)
	
	// Verify block types
	_, ok1 := message.Blocks.BlockSet[0].(*slack.SectionBlock)
	assert.True(t, ok1)
	
	_, ok2 := message.Blocks.BlockSet[1].(*slack.DividerBlock)
	assert.True(t, ok2)
	
	_, ok3 := message.Blocks.BlockSet[2].(*slack.ContextBlock)
	assert.True(t, ok3)
}

func TestBuilder_Clear(t *testing.T) {
	builder := New()
	builder.
		SetTitle("Title").
		SetClose("Close").
		SetSubmit("Submit").
		AddHeader("Header").
		AddSection("Section")
	
	// Verify builder has content
	assert.NotNil(t, builder.title)
	assert.NotNil(t, builder.close)
	assert.NotNil(t, builder.submit)
	assert.Len(t, builder.blocks, 2)
	
	// Clear the builder
	builder.Clear()
	
	// Verify everything is cleared
	assert.Nil(t, builder.title)
	assert.Nil(t, builder.close)
	assert.Nil(t, builder.submit)
	assert.Len(t, builder.blocks, 0)
}

func TestBuilder_ChainedOperations(t *testing.T) {
	modal := New().
		SetTitle("Complex Modal").
		SetClose("Cancel").
		SetSubmit("Submit").
		AddHeader("📝 User Registration").
		AddEmailInput("email", "Email Address", "", "user@example.com", false).
		AddTextInput(TextInput{
			Label:       "Full Name",
			BlockID:     "name",
			ActionID:    "name_input",
			Optional:    false,
		}).
		AddDivider().
		AddRadioButtons("role", "Select Role", "role_select", []RadioOption{
			{Text: "Admin", Value: "admin"},
			{Text: "User", Value: "user"},
		}, false).
		AddCheckboxes("permissions", "Permissions", "perm_check", []SelectOption{
			{Text: "Read", Value: "read"},
			{Text: "Write", Value: "write"},
			{Text: "Delete", Value: "delete"},
		}, true).
		BuildModal("registration", "", "")
	
	assert.Equal(t, "Complex Modal", modal.Title.Text)
	assert.Len(t, modal.Blocks.BlockSet, 6) // Header + Email + Name + Divider + Radio + Checkbox
}