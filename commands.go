package main

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/travis-ci/imaged/rpc/images"
	"golang.org/x/sync/semaphore"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
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

	sort.Sort(ByTimestamp(images))

	var b strings.Builder
	for _, image := range images {
		fmt.Fprintf(&b, "\n• `%s`", image.Name())
	}

	ReplyTo(conv).Text(b.String()).Send()
}

// RestoreBackup copies a backup image into the place of a production base image.
func RestoreBackup(ctx context.Context, conv Conversation) {
	image := conv.String("image")
	ReplyTo(conv).
		AttachText("Restoring backup for <@%s>…", conv.User()).
		Field("Image", image).
		Send()

	if err := backend.RestoreBackup(ctx, image); err != nil {
		ReplyTo(conv).ErrorText("I couldn't restore that backup.").Error(err).Send()
		return
	}

	ReplyTo(conv).
		AttachText("Successfully restored backup for <@%s>!", conv.User()).
		Color("good").
		Field("Image", image).
		Send()
}

// LastImageBuild shows information about the most recent build of an image template.
func LastImageBuild(ctx context.Context, conv Conversation) {
	image := conv.String("image")

	resp, err := imagesClient.GetLastBuild(ctx, &images.GetLastBuildRequest{Name: image})
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't load build info.").Error(err).Send()
		return
	}

	msg := ReplyTo(conv)
	finished := resp.Build.FinishedAt
	if finished == 0 {
		msg.AttachText("An %s image is currently building.", resp.Build.Name).
			Footer("imaged", time.Unix(resp.Build.CreatedAt, 0))
	} else {
		t := time.Unix(finished, 0)
		msg.AttachText("The %s image was last built %s.", resp.Build.Name, humanize.Time(t)).
			Footer("imaged", t)
	}
	updateMessage(msg, resp.Build)
	msg.Send()
}

// BuildImage starts a new image build and watches as it changes.
func BuildImage(ctx context.Context, conv Conversation) {
	image := conv.String("image")

	resp, err := imagesClient.StartBuild(ctx, &images.StartBuildRequest{
		Name:     image,
		Revision: "master",
	})
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't start the build.").Error(err).Send()
		return
	}

	build := resp.Build

	msg := ReplyTo(conv).
		AttachText("Building %s image for <@%s>…", build.Name, conv.User())
	updateMessage(msg, build)
	msg.Send()

	// Check on the build every few seconds and update the Slack message
	for {
		r, err := imagesClient.GetBuild(ctx, &images.GetBuildRequest{Id: build.Id})
		if err != nil {
			log.Printf("failed to get build info while watching build: %v", err)
		} else {
			build = r.Build
			if buildFinished(build) {
				break
			}

			updateMessage(msg, build)
			msg.Send()
		}

		time.Sleep(5 * time.Second)
	}

	// Send new message when build completes to trigger a notification
	msg = ReplyTo(conv)
	if build.Status == images.Build_SUCCEEDED {
		msg.AttachText("Successfully built %s image for <@%s>", build.Name, conv.User())
	} else {
		msg.AttachText("Failed to build %s image for <@%s>", build.Name, conv.User())
	}
	updateMessage(msg, build)
	msg.Send()
}

func buildFinished(b *images.Build) bool {
	return b.Status == images.Build_SUCCEEDED || b.Status == images.Build_FAILED
}

func buildStatus(s images.Build_Status) string {
	switch s {
	case images.Build_CREATED:
		return "Waiting to start"
	case images.Build_STARTED:
		return "Building"
	case images.Build_SUCCEEDED:
		return "Succeeded"
	case images.Build_FAILED:
		return "Failed"
	default:
		return "Unknown"
	}
}

const templatesURL = "https://github.com/travis-ci/packer-templates-mac"

func githubTreeURL(rev string) string {
	return templatesURL + "/tree/" + rev
}

func githubTreeLink(rev, display string) string {
	if display == "" {
		display = rev
	}

	return "<" + githubTreeURL(rev) + "|" + display + ">"
}

func updateMessage(msg *MessageBuilder, b *images.Build) {
	revision := b.FullRevision
	if len(revision) > 7 {
		revision = revision[0:7]
	}

	msg.
		ClearFields().
		ShortField("Build ID", strconv.FormatInt(b.Id, 10)).
		ShortField("Status", buildStatus(b.Status)).
		ShortField("Branch", githubTreeLink(b.Revision, ""))

	if revision != "" {
		msg.ShortField("Revision", githubTreeLink(b.FullRevision, revision))
	}

	switch b.Status {
	case images.Build_SUCCEEDED:
		msg.Color("good")
	case images.Build_FAILED:
		msg.Color("danger")
	default:
		msg.Color("")
	}
}

// ByTimestamp wraps a slice of Images and defines them to be sorted by the timestamp
// in their names.
type ByTimestamp []Image

func (a ByTimestamp) Len() int {
	return len(a)
}

func (a ByTimestamp) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByTimestamp) Less(i, j int) bool {
	return extractTimestamp(a[i]) < extractTimestamp(a[j])
}

func extractTimestamp(i Image) string {
	name := i.Name()
	lastDash := strings.LastIndex(name, "-")
	return name[lastDash+1 : len(name)]
}
