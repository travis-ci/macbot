package main

import (
	"context"
	"github.com/nlopes/slack"
	"github.com/travis-ci/vsphere-images"
)

func IsHostCheckedOut(ctx context.Context, msg *slack.MessageEvent) {
	typing(ctx)

	isCheckedOut, err := vsphereimages.IsHostCheckedOut(context.TODO(), vSphereURL, vSphereInsecure, packerClusterPath)
	if err != nil {
		reply(ctx, ":exclamation: Oops! I couldn't determine if a host is checked out already. `%s`", err)
		return
	}

	if isCheckedOut {
		reply(ctx, ":white_check_mark: There is a host currently checked out for building images.")
	} else {
		reply(ctx, ":heavy_multiplication_x: There is no host checked out for building images.")
	}
}

func CheckOutHost(ctx context.Context, msg *slack.MessageEvent) {
	typing(ctx)

	isCheckedOut, err := vsphereimages.IsHostCheckedOut(ctx, vSphereURL, vSphereInsecure, packerClusterPath)
	if err != nil {
		reply(ctx, ":exclamation: Oops! I couldn't determine if a host is currently checked out. `%s`", err)
		return
	}

	if isCheckedOut {
		reply(ctx, "Looks like there's already a host checked out for building images!")
		return
	}

	// Choosing a host can take a little time, so this message makes the bot more responsive
	reply(ctx, "Choosing a host to check out…")

	host, err := vsphereimages.SelectAvailableHost(ctx, vSphereURL, vSphereInsecure, prodClusterPath)
	if err != nil {
		reply(ctx, ":exclamation: Oops! I couldn't choose a host to check out. `%s`", err)
		return
	}

	// Similarly, actually checking out the host takes forever!
	// Half the point of doing this with a bot is so you can get notified when it's done after
	// you inevitably step away from your machine.
	reply(ctx, "Checking out :desktop_computer: %s for image building. I'll let you know when it's ready!", host.Name())

	err = vsphereimages.CheckOutSelectedHost(ctx, vSphereURL, vSphereInsecure, host, packerClusterPath, newProgressLogger())
	if err != nil {
		reply(ctx, ":exclamation: Oops! I couldn't check out the host. `%s`", err)
		return
	}

	reply(ctx, ":white_check_mark: Done! :desktop_computer: %s is checked out for image building.", host.Name())
}

func CheckInHost(ctx context.Context, msg *slack.MessageEvent) {
	typing(ctx)

	isCheckedOut, err := vsphereimages.IsHostCheckedOut(ctx, vSphereURL, vSphereInsecure, packerClusterPath)
	if err != nil {
		reply(ctx, ":exclamation: Oops! I couldn't determine if a host is currently checked out. `%s`", err)
		return
	}

	if !isCheckedOut {
		reply(ctx, "Looks like there isn't a host checked out right now!")
		return
	}

	reply(ctx, "Checking the host back into the production cluster...")
	host, err := vsphereimages.CheckInHost(ctx, vSphereURL, vSphereInsecure, packerClusterPath, prodClusterPath, newProgressLogger())
	if err != nil {
		reply(ctx, ":exclamation: Oops! I couldn't check the host back in. `%s`", err)
		return
	}

	reply(ctx, ":white_check_mark: Done! :desktop_computer: %s is back in production.", host.Name())
}
