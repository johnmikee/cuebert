package shared

import (
	"fmt"
	"time"

	"github.com/shomali11/slacker/v2"
	"github.com/slack-go/slack"
)

type ExclusionRequest struct {
	TriggerID  string
	Reason     string
	ReasonHint string
	Blocks     []*slack.InputBlock
}

// exclusionRequest is the modal the user will see when they request an exclusion
// the results of this modal will be sent to exclusionRequestDecision
func (c *Shared) ExclusionRequest(e *ExclusionRequest) {
	headerSection := headerBlock("Request an Exclusion")

	reasonBlock := reasonBlock(e.Reason, e.ReasonHint, ExclusionReason)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			reasonBlock,
		},
	}

	modalRequest := slack.ModalViewRequest{
		Type:       slack.VTModal,
		Title:      TitleText,
		Blocks:     blocks,
		Close:      CloseText,
		Submit:     SubmitText,
		CallbackID: c.constFormat(ExclusionModal),
	}

	_, err := c.bot.Bot().SlackClient().OpenView(e.TriggerID, modalRequest)
	if err != nil {
		c.log.Err(err).Send()
	}
}

// ExclusionRequested handles the request for an exclusion allowing authorized users to approve or deny
func (c *Shared) ExclusionRequested(ctx *slacker.InteractionContext) {
	action := ctx.Callback().ActionCallback.AttachmentActions[0]

	_, _, err := c.bot.Bot().SlackClient().DeleteMessage(ctx.Callback().Channel.ID, ctx.Callback().MessageTs)
	if err != nil {
		c.log.Info().AnErr("error", err).Msg("deleting exclusion request message")
	}

	if action.Name == NoExclusion {
		return
	}

	c.log.Info().Str("action", action.Name).Msg("exclusion requested")
	switch action.Name {
	case YesExclusion:
		c.caller.ExclusionYes(ctx)
	case ExclusionAddYes:
		c.ExclusionAdd(ctx.Callback().TriggerID)
	}
}

func (c *Shared) ExclusionAdd(triggerID string) {
	headerSection := headerBlock("Request an Exclusion")

	reasonBlock := reasonBlock("Reason for Exclusion", "ex: top-secret", c.constFormat(ExclusionReason))

	today := time.Now().Format("2006-01-02")
	datePickerBlock := datePickerBlock(
		fmt.Sprintf("ex: %s", today),
		c.constFormat(ExclusionDatePicker),
	)

	base := []slack.Block{
		headerSection,
	}
	for _, b := range c.caller.ExclusionAddBlocks() {
		base = append(base, b)
	}
	base = append(base, reasonBlock, datePickerBlock)

	blocks := slack.Blocks{
		BlockSet: base,
	}

	modalRequest := slack.ModalViewRequest{
		Type:       slack.VTModal,
		Title:      TitleText,
		Close:      CloseText,
		Submit:     SubmitText,
		Blocks:     blocks,
		CallbackID: c.constFormat(AdminExclusionModal),
	}

	_, err := c.bot.Bot().SlackClient().OpenView(triggerID, modalRequest)
	if err != nil {
		c.log.Err(err).Send()
	}
}
