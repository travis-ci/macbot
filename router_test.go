package main

import (
	"context"
	"testing"
)

func TestEmptyRouterKnowsNothing(t *testing.T) {
	router := NewRouter()
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	if reply.text != "Sorry, <@user>! I don't know how to answer that." {
		t.Fatal("unexpected reply, got", reply)
	}
}

func TestRouterSingleUnknownCommand(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("other command", func(_ context.Context, conv Conversation) {
		ReplyTo(conv).Text("Other command").Send()
	})
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	if reply.text != "Sorry, <@user>! I don't know how to answer that." {
		t.Fatal("unexpected reply, got", reply)
	}
}

func TestRouterSingleMatchingCommand(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("some command", func(_ context.Context, conv Conversation) {
		ReplyTo(conv).Text("Some command").Send()
	})
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	if reply.text != "Some command" {
		t.Fatal("unexpected reply, got", reply)
	}
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
	if reply.text != "Some command" {
		t.Fatal("unexpected reply, got", reply)
	}
}

func TestRouterIrrelevantMessage(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("some command", func(_ context.Context, conv Conversation) {
		ReplyTo(conv).Text("Some command").Send()
	})
	conv := newTestConversation("")
	router.Reply(context.TODO(), conv)

	if len(conv.replies) > 0 {
		t.Fatal("expected no replies to irrelevant message")
	}
}
