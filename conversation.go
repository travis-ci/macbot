package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"strings"
)

// Conversation provides an easy way for commands to respond to the user.
type Conversation interface {
	User() string
	Channel() string
	CommandText() string
	IsDirectMessage() bool
	Reply(string, ...interface{})
	ReplyWithError(string, error)
	ReplyWithOptions(...slack.MsgOption) string
}

type slackConversation struct {
	Event *slack.MessageEvent
}

// NewConversation creates a conversation from an incoming Slack message.
func NewConversation(event *slack.MessageEvent) Conversation {
	return &slackConversation{
		Event: event,
	}
}

// User returns the ID of the user who initiated the conversation.
func (c *slackConversation) User() string {
	return c.Event.User
}

// Channel returns the ID of the channel where the conversation is happening.
func (c *slackConversation) Channel() string {
	return c.Event.Channel
}

// CommandText extracts the command out of the message text.
//
// If the message was not directed to the bot, either through DM or an @mention,
// CommandText returns an empty string.
func (c *slackConversation) CommandText() string {
	userID := rtm.GetInfo().User.ID
	text := strings.TrimSpace(c.Event.Text)

	mentionPrefix := "<@" + userID + "> "
	if !c.IsDirectMessage() && !strings.HasPrefix(text, mentionPrefix) {
		return ""
	}

	if strings.HasPrefix(text, mentionPrefix) {
		text = text[len(mentionPrefix):len(text)]
	}

	return strings.ToLower(strings.TrimSpace(text))
}

// IsDirectMessage returns true if the conversation was started via direct message.
func (c *slackConversation) IsDirectMessage() bool {
	return strings.HasPrefix(c.Channel(), "D")
}

// Reply sends a basic text reply to the conversation.
//
// If the conversation is not a DM, the message will start with an @mention of the
// user who initiated the conversation.
func (c *slackConversation) Reply(text string, args ...interface{}) {
	text = fmt.Sprintf(text, args...)

	if !c.IsDirectMessage() {
		text = fmt.Sprintf("<@%s>: %s", c.User(), text)
	}

	m := rtm.NewOutgoingMessage(text, c.Channel())
	rtm.SendMessage(m)
}

// ReplyWithError sends a reply message indicating that an error occurred.
//
// The message will be prefixed with an apology directed at the user who initiated
// the conversation. If err is not nil, a code block will be included at the end of
// the message to show the string representation of the error.
func (c *slackConversation) ReplyWithError(text string, err error) {
	if err != nil {
		text = fmt.Sprintf("%s\n```%s```", text, err)
	}

	attachment := slack.Attachment{
		Text:  fmt.Sprintf("Sorry, <@%s>! %s", c.User(), text),
		Color: "danger",
	}
	c.ReplyWithOptions(slack.MsgOptionAttachments(attachment))
}

// ReplyWithOptions sends a message using low-level options.
//
// Returns the timestamp of the message, with can be used later to update the message.
func (c *slackConversation) ReplyWithOptions(options ...slack.MsgOption) string {
	// Always default to sending as the bot user.
	// Without this option, messages show up as being from "bot" instead of "macbot."
	options = append([]slack.MsgOption{
		slack.MsgOptionAsUser(true),
	}, options...)

	_, timestamp, _, _ := rtm.Client.SendMessage(c.Channel(), options...)
	return timestamp
}
