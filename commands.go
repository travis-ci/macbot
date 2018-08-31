package main

import (
	"context"
	"fmt"
	"github.com/nlopes/slack"
	"golang.org/x/sync/semaphore"
)

var hostSemaphore = semaphore.NewWeighted(1)

// IsHostCheckedOut checks if a host is present in the dev cluster.
func IsHostCheckedOut(ctx context.Context, msg *slack.MessageEvent) {
	isCheckedOut, err := backend.IsHostCheckedOut(ctx)
	if err != nil {
		replyError(ctx, "I couldn't determine if a host is checked out already.", err)
		return
	}

	if isCheckedOut {
		reply(ctx, ":white_check_mark: There is a host currently checked out for building images.")
	} else {
		reply(ctx, ":heavy_multiplication_x: There is no host checked out for building images.")
	}
}

// CheckOutHost chooses an available host in the production cluster and moves it to the dev
// cluster.
//
// If there is already a host in the dev cluster, it informs the user who asked. Only one user can
// attempt to check out/in a host at a time: other users will get an error message when they try.
func CheckOutHost(ctx context.Context, msg *slack.MessageEvent) {
	isCheckedOut, err := backend.IsHostCheckedOut(ctx)
	if err != nil {
		replyError(ctx, "I couldn't determine if a host is currently checked out.", err)
		return
	}

	if isCheckedOut {
		replyError(ctx, "Looks like there's already a host checked out for building images!", nil)
		return
	}

	canCheckOut := hostSemaphore.TryAcquire(1)
	if !canCheckOut {
		replyError(ctx, "Someone is already trying to check in/out a host right now, try again later!", nil)
		return
	}
	defer hostSemaphore.Release(1)

	// Choosing a host can take a little time, so this message makes the bot more responsive
	attachment := slack.Attachment{
		Text: "Choosing a host to check out for <@" + msg.User + ">…",
	}
	_, timestamp, _, _ := rtm.Client.SendMessage(msg.Channel, slack.MsgOptionAsUser(true), slack.MsgOptionAttachments(attachment))

	host, err := backend.SelectHost(ctx)
	if err != nil {
		replyError(ctx, "I couldn't choose a host to check out.", err)
		return
	}

	// Similarly, actually checking out the host takes forever!
	// Half the point of doing this with a bot is so you can get notified when it's done after
	// you inevitably step away from your machine.
	attachment.Text = fmt.Sprintf("Checking out host for <@%s>…", msg.User)
	attachment.Fields = []slack.AttachmentField{
		{
			Title: "Host",
			Value: fmt.Sprintf(":desktop_computer: %s", host.Name()),
		},
	}
	rtm.Client.SendMessage(msg.Channel, slack.MsgOptionUpdate(timestamp), slack.MsgOptionAsUser(true), slack.MsgOptionAttachments(attachment))

	err = backend.CheckOutHost(ctx, host)
	if err != nil {
		replyError(ctx, "I couldn't check out the host.", err)
		return
	}

	attachment.Text = fmt.Sprintf("Successfully checked out host for <@%s>!", msg.User)
	attachment.Color = "good"
	rtm.Client.SendMessage(msg.Channel, slack.MsgOptionAsUser(true), slack.MsgOptionAttachments(attachment))
}

// CheckInHost moves a host from the dev cluster to the production cluster.
//
// If there is no host in the dev cluster, it informs the user who asked.
func CheckInHost(ctx context.Context, msg *slack.MessageEvent) {
	isCheckedOut, err := backend.IsHostCheckedOut(ctx)
	if err != nil {
		replyError(ctx, "I couldn't determine if a host is currently checked out.", err)
		return
	}

	if !isCheckedOut {
		replyError(ctx, "Looks like there isn't a host checked out right now!", nil)
		return
	}

	canCheckOut := hostSemaphore.TryAcquire(1)
	if !canCheckOut {
		replyError(ctx, "Someone is already trying to check in/out a host right now, try again later!", nil)
		return
	}
	defer hostSemaphore.Release(1)

	attachment := slack.Attachment{
		Text: "Checking the host in for <@" + msg.User + ">…",
	}
	rtm.Client.SendMessage(msg.Channel, slack.MsgOptionAsUser(true), slack.MsgOptionAttachments(attachment))

	host, err := backend.CheckInHost(ctx)
	if err != nil {
		replyError(ctx, "I couldn't check the host back in.", err)
		return
	}

	attachment.Text = fmt.Sprintf("Successfully checked in host for <@%s>!", msg.User)
	attachment.Color = "good"
	attachment.Fields = []slack.AttachmentField{
		{
			Title: "Host",
			Value: fmt.Sprintf(":desktop_computer: %s", host.Name()),
		},
	}
	rtm.Client.SendMessage(msg.Channel, slack.MsgOptionAsUser(true), slack.MsgOptionAttachments(attachment))
}
