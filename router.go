package main

import (
	"context"
)

// Router dispatches conversations to handler functions.
type Router struct {
	handlers map[string]HandlerFunc
}

// HandlerFunc is a function that can reply to a conversation.
type HandlerFunc func(context.Context, Conversation)

// NewRouter creates a new router with no handlers registered.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]HandlerFunc),
	}
}

// HandleFunc registers a function as a handler for a command.
func (r *Router) HandleFunc(command string, fn HandlerFunc) {
	r.handlers[command] = fn
}

// Reply sends a conversation to a registered handler if one matches.
// If no handler matches, Reply will send an error reply message to the conversation.
// If the command text is an empty string, Reply will ignore the message.
func (r *Router) Reply(ctx context.Context, conv Conversation) {
	text := conv.CommandText()
	if text == "" {
		return
	}

	if fn, ok := r.handlers[text]; ok {
		fn(ctx, conv)
	} else {
		unknownCommand(ctx, conv)
	}
}

func unknownCommand(ctx context.Context, conv Conversation) {
	conv.ReplyWithError("I don't know how to answer that.", nil)
}
