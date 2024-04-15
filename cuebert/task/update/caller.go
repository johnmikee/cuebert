package update

import (
	"fmt"
	"time"

	"github.com/johnmikee/cuebert/cuebert/task/shared"
	"github.com/johnmikee/cuebert/db/shared/users"
)

func (c *Client) FirstMessageSent(serviceID, slackID string) error {
	return c.common.FirstMessageSent(serviceID, slackID)
}

// FirstMessageDelivered implements task.Rabbit.
func (c *Client) FirstMessageDelivered(serviceID string, slackID string) error {
	return c.common.FirstMessageDelivered(serviceID, slackID)
}

func (c *Client) ReminderMessage(r *shared.ReminderPayload) error {
	return c.common.ReminderMessage(r)
}

// MessageAcknowledged runs when the message has been acknowledged
func (c *Client) MessageAcknowledged(userID string, when time.Time, which string) error {
	return c.common.MessageAcknowledged(userID, when, which)
}

// UserCompliant implements task.Rabbit.
func (c *Client) UserCompliant(userID string) error {
	ui, err := c.tables.User().By().Opts(users.UserSlackID, userID).Query()
	if err != nil {
		return err
	}

	if ui.Empty() {
		return fmt.Errorf("user not found")
	}
	return c.idp.RemoveFromBlocked(ui[0].UserEmail)
}
