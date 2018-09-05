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
	if reply != "I don't know how to answer that." {
		t.Fatal("unexpected reply, got", reply)
	}
}

func TestRouterSingleUnknownCommand(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("other command", func(_ context.Context, conv Conversation) {
		conv.Reply("Other command")
	})
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	if reply != "I don't know how to answer that." {
		t.Fatal("unexpected reply, got", reply)
	}
}

func TestRouterSingleMatchingCommand(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("some command", func(_ context.Context, conv Conversation) {
		conv.Reply("Some command")
	})
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	if reply != "Some command" {
		t.Fatal("unexpected reply, got", reply)
	}
}

func TestRouterMultipleCommands(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("other command", func(_ context.Context, conv Conversation) {
		conv.Reply("Other command")
	})
	router.HandleFunc("some command", func(_ context.Context, conv Conversation) {
		conv.Reply("Some command")
	})
	conv := newTestConversation("some command")
	router.Reply(context.TODO(), conv)

	reply := conv.replies[0]
	if reply != "Some command" {
		t.Fatal("unexpected reply, got", reply)
	}
}

func TestRouterIrrelevantMessage(t *testing.T) {
	router := NewRouter()
	router.HandleFunc("some command", func(_ context.Context, conv Conversation) {
		conv.Reply("Some command")
	})
	conv := newTestConversation("")
	router.Reply(context.TODO(), conv)

	if len(conv.replies) > 0 {
		t.Fatal("expected no replies to irrelevant message")
	}
}
