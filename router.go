package main

import (
	"context"
	"fmt"
	"github.com/shomali11/commander"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Router dispatches conversations to handler functions.
type Router struct {
	commands []command
}

type command struct {
	*commander.Command

	pattern string
	handler HandlerFunc
}

// HandlerFunc is a function that can reply to a conversation.
type HandlerFunc func(context.Context, Conversation)

// NewRouter creates a new router with no handlers registered.
func NewRouter() *Router {
	return &Router{}
}

// HandleFunc registers a function as a handler for a command.
func (r *Router) HandleFunc(pattern string, fn HandlerFunc) {
	cmd := commander.NewCommand(pattern)
	r.commands = append(r.commands, command{
		Command: cmd,
		pattern: pattern,
		handler: fn,
	})
	log.WithField("pattern", pattern).Debug("added command to router")
}

// Reply sends a conversation to a registered handler if one matches.
// If no handler matches, Reply will send an error reply message to the conversation.
// If the command text is an empty string, Reply will ignore the message.
func (r *Router) Reply(ctx context.Context, conv Conversation) {
	entry := log.WithFields(log.Fields{
		"user":    conv.User(),
		"channel": conv.Channel(),
	})

	text := conv.CommandText()
	if text == "" {
		entry.Debug("ignoring irrelevant message")
		return
	}

	entry = entry.WithField("command", text)

	for _, c := range r.commands {
		if props, ok := c.Match(text); ok {
			conv.SetProperties(props)
			entry.WithField("pattern", c.pattern).Info("handling command")
			c.handler(ctx, conv)
			return
		}
	}

	if text == "help" {
		entry.Info("sending help")
		r.help(ctx, conv)
	} else {
		entry.Warn("handling unknown command")
		r.unknownCommand(ctx, conv)
	}
}

func (r *Router) unknownCommand(ctx context.Context, conv Conversation) {
	var b strings.Builder
	b.WriteString("I don't know how to answer that.")

	if len(r.commands) > 0 {
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
		fmt.Fprintf(&b, "• `%s`\n", cmd.pattern)
	}

	return b.String()
}
