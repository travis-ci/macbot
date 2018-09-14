package main

import (
	"github.com/shomali11/proper"
	"strconv"
	"strings"
)

type testConversation struct {
	channel   string
	user      string
	command   string
	replies   []MessageBuilder
	timestamp int

	*proper.Properties
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
	return strings.HasPrefix(c.channel, "D")
}

func (c *testConversation) Send(b *MessageBuilder) string {
	c.timestamp++
	timestamp := strconv.Itoa(c.timestamp)
	c.replies = append(c.replies, *b)
	return timestamp
}

func (c *testConversation) SetProperties(props *proper.Properties) {
	c.Properties = props
}

func (c *testConversation) String(key string) string {
	return c.StringParam(key, "")
}
