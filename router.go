package main

import (
	"context"
	"fmt"
	"strings"
)

// Router dispatches conversations to handler functions.
type Router struct {
	handlers map[string]HandlerFunc
	commands []string
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
	r.commands = append(r.commands, command)
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
	} else if text == "help" {
		r.help(ctx, conv)
	} else {
		r.unknownCommand(ctx, conv)
	}
}

func (r *Router) unknownCommand(ctx context.Context, conv Conversation) {
	var b strings.Builder
	b.WriteString("I don't know how to answer that.")

	if len(r.handlers) > 0 {
		b.WriteString(" I can respond to the following commands:\n\n")
		b.WriteString(r.commandList())
	}

	ReplyTo(conv).ErrorText(b.String()).Send()
}

func (r *Router) help(ctx context.Context, conv Conversation) {
	ReplyTo(conv).Text("\n" + r.commandList()).Send()
}

func (r *Router) commandList() string {
	var b strings.Builder

	for _, cmd := range r.commands {
		fmt.Fprintf(&b, "• `%s`\n", cmd)
	}

	return b.String()
}
