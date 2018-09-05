package main

import (
	"fmt"
	"github.com/nlopes/slack"
)

type testConversation struct {
	channel string
	user    string
	command string
	replies []string
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

func (c *testConversation) Reply(text string, args ...interface{}) {
	msg := fmt.Sprintf(text, args...)
	c.replies = append(c.replies, msg)
}

func (c *testConversation) ReplyWithError(text string, err error) {
	// TODO: include err information
	c.replies = append(c.replies, text)
}

func (c *testConversation) ReplyWithOptions(options ...slack.MsgOption) string {
	// TODO: implement this or change the interface to something easier to reason about
	return ""
}
