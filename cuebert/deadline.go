package main

import (
	"fmt"
	"time"
)

func (c *Cuebert) checkDeadline(time.Time) {
	now := time.Now()
	deadline := fmt.Sprintf("%s %s", c.flags.deadline, c.flags.cutoffTime)

	layout := "01-02-2006 15:04"
	t, err := time.Parse(layout, deadline)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	if now.After(t) {
		c.method.Deadline()
	}
	c.stop()
}
