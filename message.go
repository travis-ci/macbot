package main

import (
	"fmt"
	"strings"
)

// MessageBuilder provides a fluent interface for constructing messages to
// send as replies to a conversation.
type MessageBuilder struct {
	conversation Conversation
	text         string
	error        error
	isAttachment bool
	color        string
	fields       []messageField
	timestamp    string
}

type messageField struct {
	title string
	value string
}

// ReplyTo creates a MessageBuilder for a conversation, starting with an
// empty message.
func ReplyTo(c Conversation) *MessageBuilder {
	return &MessageBuilder{
		conversation: c,
	}
}

// Text sets the text of the message, using a Printf-style format string.
// If the conversation is not a direct message, an @mention of the user who
// started the conversation will be included at the start of the text.
func (b *MessageBuilder) Text(text string, args ...interface{}) *MessageBuilder {
	b.text = fmt.Sprintf(text, args...)

	if !b.conversation.IsDirectMessage() && !strings.Contains(b.text, "<@"+b.conversation.User()+">") {
		b.text = fmt.Sprintf("<@%s>: %s", b.conversation.User(), b.text)
	}

	return b
}

// ErrorText sets the text of the message for sending an error to the user.
// ErrorText uses a Printf-style format string. The message will be send with an
// attachment in a red color, and will include an apologetic prefix on the message
// text.
func (b *MessageBuilder) ErrorText(text string, args ...interface{}) *MessageBuilder {
	userText := fmt.Sprintf(text, args...)
	b.text = fmt.Sprintf("Sorry, <@%s>! %s", b.conversation.User(), userText)
	b.isAttachment = true
	b.color = "danger"
	return b
}

// Error attaches a Go error to the message.
// The string representation of the error will be appended to the message in a code block.
func (b *MessageBuilder) Error(err error) *MessageBuilder {
	b.error = err
	return b
}

// AttachText sets the text of the message and forces it to be sent as an attachment.
// This can be used to get the formatting of an attachment even if the message doesn't
// otherwise use attachment features like fields.
func (b *MessageBuilder) AttachText(text string, args ...interface{}) *MessageBuilder {
	b.isAttachment = true
	return b.Text(text, args...)
}

// Color sets the color of the attachment for the message.
// This forces the message to be sent as an attachment.
func (b *MessageBuilder) Color(c string) *MessageBuilder {
	b.color = c
	b.isAttachment = true
	return b
}

// Field adds a field to the attachment for the message.
// This forces the message to be sent as an attachment.
func (b *MessageBuilder) Field(title string, text string, args ...interface{}) *MessageBuilder {
	b.isAttachment = true
	b.fields = append(b.fields, messageField{
		title: title,
		value: fmt.Sprintf(text, args...),
	})
	return b
}

// Send sends the message as a reply to the conversation.
// The message builder keeps track of the timestamp of the message, allowing
// the same builder to be used to later update the message.
func (b *MessageBuilder) Send() *MessageBuilder {
	b.timestamp = b.conversation.Send(b)
	return b
}
