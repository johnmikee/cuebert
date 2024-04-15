package update

// MessageTest implements task.Rabbit.
func (c *Client) MessageTest(which string) {
	panic("unimplemented")
}

// FirstMessage implements task.Rabbit.
func (c *Client) FirstMessage() string {
	return c.impl.FirstMessage()
}

// FirstMessageHeader implements task.Rabbit.
func (c *Client) FirstMessageHeader() string {
	return c.impl.FirstMessageHeader()
}

// FinalMessage implements task.Rabbit.
func (c *Client) FinalMessage() string {
	return c.impl.FinalMessage()
}
