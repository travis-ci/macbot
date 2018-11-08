package main

import (
	"context"
	"github.com/dustin/go-humanize"
	log "github.com/sirupsen/logrus"
	"github.com/travis-ci/imaged/rpc/images"
	"time"
)

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
	branch := conv.String("branch")
	if branch == "" {
		branch = "master"
	}

	resp, err := imagesClient.StartBuild(ctx, &images.StartBuildRequest{
		Name:     image,
		Revision: branch,
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

func buildStatus(b *images.Build) string {
	switch b.Status {
	case images.Build_CREATED:
		return "Waiting to start"
	case images.Build_STARTED:
		return "Building"
	case images.Build_SUCCEEDED:
		return buildLogLink(b, "Succeeded")
	case images.Build_FAILED:
		return buildLogLink(b, "Failed")
	default:
		return "Unknown"
	}
}

func buildLogLink(b *images.Build, text string) string {
	resp, err := imagesClient.GetRecordURL(context.Background(), &images.GetRecordURLRequest{
		BuildId:  b.Id,
		FileName: "build.log",
	})
	if err != nil {
		return text
	}

	return "<" + resp.Url + "|" + text + ">"
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
		ShortField("Build ID", "%d", b.Id).
		ShortField("Status", "%s", buildStatus(b)).
		ShortField("Branch", "%s", githubTreeLink(b.Revision, ""))

	if revision != "" {
		msg.ShortField("Revision", "%s", githubTreeLink(b.FullRevision, revision))
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
