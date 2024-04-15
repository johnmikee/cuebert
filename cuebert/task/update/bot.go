package update

import (
	"strings"

	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

// ExclusionRequestDecision executes the exclusion request decision helper
func (c *Client) ExclusionRequestDecision(ctx *slacker.InteractionContext) {
	c.common.ExclusionRequestDecision(ctx)
}

// ExclusionYes runs when an exclusion is approved
func (c *Client) ExclusionYes(ctx *slacker.InteractionContext) {
	c.common.ExclusionYes(ctx)
}

// ExclusionAddBlocks builds the exclusion add blocks.
func (c *Client) ExclusionAddBlocks() []*slack.InputBlock {
	return c.common.ExclusionAddBlocks()
}

// ExclusionSubmit submits the exclusion request
func (c *Client) ExclusionSubmit(ctx *slacker.InteractionContext) {
	c.common.ExclusionSubmit(ctx)
}

// GetUserTZ implements task.Rabbit getting the user's timezone.
func (c *Client) GetUserTZ(slackID string) (int64, error) {
	return c.common.GetUserTZ(slackID)
}

// UserExclusionSubmit submits the user exclusion request
func (c *Client) UserExclusionSubmit(ctx *slacker.InteractionContext) {
	c.common.UserExclusionSubmit(ctx)
}

// UpdateReminderTime implements task.Rabbit updating the reminder time.
func (c *Client) UpdateReminderTime(dateVal string, timeVal string, slackID string) error {
	return c.common.UpdateReminderTime(dateVal, timeVal, slackID)
}

// UserUpdate runs the user update modal
func (c *Client) UserUpdate(ctx *slacker.InteractionContext) {
	c.common.UserUpdate(ctx)
}

// UserUpdateSubmit submits the user update request
func (c *Client) UserUpdateSubmit(ctx *slacker.InteractionContext) {
	c.common.UserUpdateSubmit(ctx)
}

func (c *Client) InteractionHandler(ctx *slacker.InteractionContext) {
	c.shared.InteractionHandler(ctx)
}

func (c *Client) InteractiveHelper(ctx *slacker.InteractionContext) {
	c.shared.InteractiveHelper(ctx)
}

func (c *Client) LoadInput(ctx *slacker.InteractionContext, loadType string) {
	c.shared.LoadInput(ctx, loadType)
}

func (c *Client) LoadProgram(triggerID string, loadType string) {
	c.shared.LoadProgram(
		&shared.Loader{
			TriggerID: triggerID,
			LoadType:  loadType,
			Tables:    shared.Tables(c.config.Flags.TaskImplementation, shared.TaskType(strings.ToLower(c.config.Flags.Task))),
		},
	)
}

func (c *Client) ModalSubmit(ctx *slacker.InteractionContext) {
	c.shared.ModalSubmit(ctx)
}

func (c *Client) RegisterHelp() {
	c.common.RegisterHelp()
}

func (c *Client) RegisterCommands() {
	c.shared.ExclusionHelp()
	c.shared.ReminderHelp()
	c.common.GetSelfInfo()
	c.common.DeviceInfo()
	c.shared.CheckComplianceStatus()
}

func (c *Client) RegisterAdminCommands() {
	c.common.GetUsersInfo()
	c.shared.AddExclusion()
	c.shared.InitSettings()
	c.shared.Stopper()
	c.shared.UpdateConfig()
	c.shared.GetStatus(c.config.StatusHandler)
}

func (c *Client) RegisterMiddlewares() {
	c.shared.RegisterMiddlewares()
}

func (c *Client) RegisterHandlers() {
	c.shared.RegisterHandlers()
}
