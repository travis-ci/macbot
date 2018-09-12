package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"strings"
)

var hostSemaphore = semaphore.NewWeighted(1)

// IsHostCheckedOut checks if a host is present in the dev cluster.
func IsHostCheckedOut(ctx context.Context, conv Conversation) {
	isCheckedOut, err := backend.IsHostCheckedOut(ctx)
	if err != nil {
		ReplyTo(conv).
			ErrorText("I couldn't determine if a host is checked out already.").
			Error(err).Send()
		return
	}

	if isCheckedOut {
		ReplyTo(conv).Text(":white_check_mark: There is a host currently checked out for building images.").Send()
	} else {
		ReplyTo(conv).Text(":heavy_multiplication_x: There is no host checked out for building images.").Send()
	}
}

// CheckOutHost chooses an available host in the production cluster and moves it to the dev
// cluster.
//
// If there is already a host in the dev cluster, it informs the user who asked. Only one user can
// attempt to check out/in a host at a time: other users will get an error message when they try.
func CheckOutHost(ctx context.Context, conv Conversation) {
	isCheckedOut, err := backend.IsHostCheckedOut(ctx)
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't determine if a host is currently checked out.").Error(err).Send()
		return
	}

	if isCheckedOut {
		ReplyTo(conv).ErrorText("Looks like there's already a host checked out for building images!").Send()
		return
	}

	canCheckOut := hostSemaphore.TryAcquire(1)
	if !canCheckOut {
		ReplyTo(conv).ErrorText("Someone is already trying to check in/out a host right now, try again later!").Send()
		return
	}
	defer hostSemaphore.Release(1)

	// Choosing a host can take a little time, so this message makes the bot more responsive
	msg := ReplyTo(conv).AttachText("Choosing a host to check out for <@%s>…", conv.User()).Send()

	host, err := backend.SelectHost(ctx)
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't choose a host to check out.").Error(err).Send()
		return
	}

	// Similarly, actually checking out the host takes forever!
	// Half the point of doing this with a bot is so you can get notified when it's done after
	// you inevitably step away from your machine.
	msg.AttachText("Checking out host for <@%s>…", conv.User()).
		Field("Host", ":desktop_computer: %s", host.Name()).
		Send()

	err = backend.CheckOutHost(ctx, host)
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't check out the host.").Error(err).Send()
		return
	}

	ReplyTo(conv).
		Text("Successfully checked out host for <@%s>!", conv.User()).
		Field("Host", ":desktop_computer: %s", host.Name()).
		Color("good").Send()
}

// CheckInHost moves a host from the dev cluster to the production cluster.
//
// If there is no host in the dev cluster, it informs the user who asked.
func CheckInHost(ctx context.Context, conv Conversation) {
	isCheckedOut, err := backend.IsHostCheckedOut(ctx)
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't determine if a host is currently checked out.").Error(err).Send()
		return
	}

	if !isCheckedOut {
		ReplyTo(conv).ErrorText("Looks like there isn't a host checked out right now!").Send()
		return
	}

	canCheckOut := hostSemaphore.TryAcquire(1)
	if !canCheckOut {
		ReplyTo(conv).ErrorText("Someone is already trying to check in/out a host right now, try again later!").Send()
		return
	}
	defer hostSemaphore.Release(1)

	ReplyTo(conv).AttachText("Checking the host in for <@%s>…", conv.User()).Send()

	host, err := backend.CheckInHost(ctx)
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't check the host back in.").Error(err).Send()
		return
	}

	ReplyTo(conv).
		AttachText("Successfully checked in host for <@%s>!", conv.User()).
		Color("good").
		Field("Host", ":desktop_computer: %s", host.Name()).
		Send()
}

// BaseImages lists the names of the base VM images that are in the datacenter.
func BaseImages(ctx context.Context, conv Conversation) {
	images, err := backend.BaseImages(ctx)
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't get the list of base images.").Error(err).Send()
		return
	}

	var b strings.Builder
	for _, image := range images {
		fmt.Fprintf(&b, "\n• `%s`", image.Name())
	}

	ReplyTo(conv).Text(b.String()).Send()
}
