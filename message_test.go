package main

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSimpleReplyDirect(t *testing.T) {
	conv := newTestConversation("foo")
	conv.channel = "D123"

	msg := ReplyTo(conv).Text("This is a message.")
	require.Equal(t, "This is a message.", msg.text)
	require.False(t, msg.isAttachment)
}

func TestSimpleReplyChannel(t *testing.T) {
	conv := newTestConversation("foo")

	msg := ReplyTo(conv).Text("This is a message.")
	require.Equal(t, "<@user>: This is a message.", msg.text)
	require.False(t, msg.isAttachment)
}

func TestSimpleReplyAlreadyMentioned(t *testing.T) {
	conv := newTestConversation("foo")

	msg := ReplyTo(conv).Text("Hey <@user>! This is a message.")
	require.Equal(t, "Hey <@user>! This is a message.", msg.text)
	require.False(t, msg.isAttachment)
}

func TestErrorReply(t *testing.T) {
	conv := newTestConversation("foo")

	msg := ReplyTo(conv).ErrorText("I made a mistake.")
	require.Equal(t, "Sorry, <@user>! I made a mistake.", msg.text)
	require.True(t, msg.isAttachment)
	require.Equal(t, "danger", msg.color)
}

func TestErrorReplyWithError(t *testing.T) {
	conv := newTestConversation("foo")

	err := errors.New("aw dang")
	msg := ReplyTo(conv).ErrorText("I made a mistake.").Error(err)
	require.Equal(t, "Sorry, <@user>! I made a mistake.", msg.text)
	require.Equal(t, err, msg.error)
	require.True(t, msg.isAttachment)
	require.Equal(t, "danger", msg.color)
}

func TestSimpleAttachmentReply(t *testing.T) {
	conv := newTestConversation("foo")

	msg := ReplyTo(conv).AttachText("It's happening!")
	require.Equal(t, "<@user>: It's happening!", msg.text)
	require.True(t, msg.isAttachment)
	require.Empty(t, msg.color)
}

func TestComplexMessage(t *testing.T) {
	conv := newTestConversation("foo")

	msg := ReplyTo(conv).
		Text("A thing happened!").
		Color("abcdef").
		Field("Foo", "abc %d", 123).
		Field("Bar", "def")

	require.Equal(t, "<@user>: A thing happened!", msg.text)
	require.True(t, msg.isAttachment)
	require.Equal(t, "abcdef", msg.color)

	require.Equal(t, "Foo", msg.fields[0].title)
	require.Equal(t, "abc 123", msg.fields[0].value)
	require.Equal(t, "Bar", msg.fields[1].title)
	require.Equal(t, "def", msg.fields[1].value)
}

func TestTimestampWhenSent(t *testing.T) {
	conv := newTestConversation("foo")

	msg := ReplyTo(conv).Text("Test message.")
	require.Empty(t, msg.timestamp)

	msg = msg.Send()
	require.NotEmpty(t, msg.timestamp)
}
