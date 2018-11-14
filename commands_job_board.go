package main

import (
	"context"
	"fmt"
	"strings"
)

// ListImages lists the images registered in job board.
func ListImages(ctx context.Context, conv Conversation) {
	env := conv.String("env")
	if env == "" {
		env = "production"
	}

	jb, found := jobBoards[env]
	if !found {
		ReplyTo(conv).ErrorText("No job board is configured for the %s environment.", env).Send()
		return
	}

	images, err := jb.ListImages(ctx)
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't get the list of images from job board.").Error(err).Send()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "macOS images registered in job-board-%s:\n", env)
	for _, i := range images {
		fmt.Fprintf(&b, "\n*%s*: `%s`", i.Tag, i.Name)
	}

	ReplyTo(conv).Text(b.String()).Send()
}

// RegisterImage adds the image to job board as a macOS build image.
func RegisterImage(ctx context.Context, conv Conversation) {
	image := conv.String("image")
	tag := conv.String("tag")
	env := conv.String("env")
	if env == "" {
		env = "production"
	}

	jb, found := jobBoards[env]
	if !found {
		ReplyTo(conv).ErrorText("No job board is configured for the %s environment.", env).Send()
		return
	}

	err := jb.RegisterImage(ctx, image, tag)
	if err != nil {
		ReplyTo(conv).ErrorText("I couldn't register the image with job board.").Error(err).Send()
		return
	}

	ReplyTo(conv).
		AttachText("Successfully registered image for <@%s>", conv.User()).
		Color("good").
		Field("Image", image).
		ShortField("Tag", tag).
		ShortField("Environment", env).
		Send()
}
