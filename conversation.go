package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/shomali11/proper"
	"strings"
)

// Conversation provides an easy way for commands to respond to the user.
type Conversation interface {
	User() string
	Channel() string
	CommandText() string
	IsDirectMessage() bool
	Send(*MessageBuilder) string

	SetProperties(*proper.Properties)
	String(string) string
}

type slackConversation struct {
	Event *slack.MessageEvent
	*proper.Properties
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
	// Ignore messages that the bot sent, no matter what
	if c.User() == userID {
		return ""
	}

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

func (c *slackConversation) Send(b *MessageBuilder) string {
	options := messageOptions(b)
	// TODO: check and probably log the error here
	_, timestamp, _, _ := rtm.Client.SendMessage(c.Channel(), options...)
	return timestamp
}

func messageOptions(b *MessageBuilder) []slack.MsgOption {
	// Always default to sending as the bot user.
	// Without this option, messages show up as being from "bot" instead of "macbot."
	options := []slack.MsgOption{
		slack.MsgOptionAsUser(true),
	}

	if b.text != "" {
		text := b.text
		if b.error != nil {
			text = fmt.Sprintf("%s\n```%s```", text, b.error)
		}

		if b.isAttachment {
			attachment := slack.Attachment{
				Text: text,
			}
			if b.color != "" {
				attachment.Color = b.color
			}
			for _, field := range b.fields {
				attachment.Fields = append(attachment.Fields, slack.AttachmentField{
					Title: field.title,
					Value: field.value,
				})
			}
			options = append(options, slack.MsgOptionAttachments(attachment))
		} else {
			options = append(options, slack.MsgOptionText(text, false))
		}
	}

	if b.timestamp != "" {
		options = append(options, slack.MsgOptionUpdate(b.timestamp))
	}

	return options
}

func (c *slackConversation) SetProperties(props *proper.Properties) {
	c.Properties = props
}

func (c *slackConversation) String(key string) string {
	return c.StringParam(key, "")
}
