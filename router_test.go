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
	require.Equal(t, "Sorry, <@user>! I don't know how to answer that.", reply.text)
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
