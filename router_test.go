package main

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEmptyRouterKnowsNothing(t *testing.T) {
	router := NewRouter()
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	require.Equal(t, "Sorry, <@user>! I don't know how to answer that.", reply.text)
}

func TestRouterSingleUnknownCommand(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("other command", func(_ context.Context, conv Conversation) {
		ReplyTo(conv).Text("Other command").Send()
	})
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	expected := "Sorry, <@user>! I don't know how to answer that. I can respond to the following commands:\n\n• `other command`\n"
	require.Equal(t, expected, reply.text)
}

func TestRouterSingleMatchingCommand(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("some command", func(_ context.Context, conv Conversation) {
		ReplyTo(conv).Text("Some command").Send()
	})
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	require.Equal(t, "<@user>: Some command", reply.text)
}

func TestRouterMultipleCommands(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("other command", func(_ context.Context, conv Conversation) {
		ReplyTo(conv).Text("Other command").Send()
	})
	router.HandleFunc("some command", func(_ context.Context, conv Conversation) {
		ReplyTo(conv).Text("Some command").Send()
	})
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	require.Equal(t, "<@user>: Some command", reply.text)
}

func TestRouterIrrelevantMessage(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("some command", func(_ context.Context, conv Conversation) {
		ReplyTo(conv).Text("Some command").Send()
	})
	conv := newTestConversation("")
	router.Reply(context.TODO(), conv)

	require.Empty(t, conv.replies)
}

func TestRouterHelp(t *testing.T) {
	router := NewRouter()
	dummy := func(_ context.Context, conv Conversation) {}
	router.HandleFunc("command 1", dummy)
	router.HandleFunc("command 2", dummy)
	router.HandleFunc("command 3", dummy)

	conv := newTestConversation("help")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	expected := "<@user>: \n• `command 1`\n• `command 2`\n• `command 3`\n"
	require.Equal(t, expected, reply.text)
}
