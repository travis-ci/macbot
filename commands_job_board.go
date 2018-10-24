package main

import (
	"context"
)

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
