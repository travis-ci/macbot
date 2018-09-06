package main

import (
	"strconv"
)

type testConversation struct {
	channel   string
	user      string
	command   string
	replies   []MessageBuilder
	timestamp int
}

func newTestConversation(command string) *testConversation {
	return &testConversation{
		channel: "test",
		user:    "user",
		command: command,
		replies: nil,
	}
}

func (c *testConversation) Channel() string {
	return c.channel
}

func (c *testConversation) User() string {
	return c.user
}

func (c *testConversation) CommandText() string {
	return c.command
}

func (c *testConversation) IsDirectMessage() bool {
	return true
}

func (c *testConversation) Send(b *MessageBuilder) string {
	c.timestamp++
	timestamp := strconv.Itoa(c.timestamp)
	c.replies = append(c.replies, *b)
	return timestamp
}
