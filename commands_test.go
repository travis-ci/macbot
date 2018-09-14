package main

import (
	"context"
	"github.com/shomali11/proper"
	"github.com/stretchr/testify/require"
	"testing"
)

func resetBackend() {
	backend = &DebugBackend{
		Host:         DebugHost("1.2.3.4"),
		disableSleep: true,
	}
}

func TestIsHostCheckedOut(t *testing.T) {
	resetBackend()
	conv := newTestConversation("is checked out")
	IsHostCheckedOut(context.TODO(), conv)

	reply := conv.replies[0]
	require.Contains(t, reply.text, "There is no host checked out for building images.")

	// now check out a host and try again
	host, _ := backend.SelectHost(context.TODO())
	backend.CheckOutHost(context.TODO(), host)

	conv = newTestConversation("is checked out")
	IsHostCheckedOut(context.TODO(), conv)

	reply = conv.replies[0]
	require.Contains(t, reply.text, "There is a host currently checked out for building images.")
}

func TestCheckOutHost(t *testing.T) {
	resetBackend()
	conv := newTestConversation("check out host")
	CheckOutHost(context.TODO(), conv)

	require.Len(t, conv.replies, 3)
	reply := conv.replies[0]
	require.Equal(t, "Choosing a host to check out for <@user>…", reply.text)
	require.True(t, reply.isAttachment)

	reply = conv.replies[1]
	require.Equal(t, "Checking out host for <@user>…", reply.text)
	f := messageField{"Host", ":desktop_computer: 1.2.3.4"}
	require.Equal(t, f, reply.fields[0])
	require.NotEmpty(t, reply.timestamp)

	reply = conv.replies[2]
	require.Equal(t, "Successfully checked out host for <@user>!", reply.text)
	require.Equal(t, f, reply.fields[0])
	require.Equal(t, "good", reply.color)
	require.Empty(t, reply.timestamp, "expected reply 2 to be its own message")

	require.True(t, backend.(*DebugBackend).isCheckedOut)
}

func TestCheckOutHostAlreadyOut(t *testing.T) {
	resetBackend()
	host, _ := backend.SelectHost(context.TODO())
	backend.CheckOutHost(context.TODO(), host)

	conv := newTestConversation("check out host")
	CheckOutHost(context.TODO(), conv)

	reply := conv.replies[0]
	require.Equal(t, "Sorry, <@user>! Looks like there's already a host checked out for building images!", reply.text)
}

func TestCheckInHost(t *testing.T) {
	resetBackend()
	host, _ := backend.SelectHost(context.TODO())
	backend.CheckOutHost(context.TODO(), host)

	conv := newTestConversation("check in host")
	CheckInHost(context.TODO(), conv)

	require.Len(t, conv.replies, 2)
	reply := conv.replies[0]
	require.Equal(t, "Checking the host in for <@user>…", reply.text)
	require.True(t, reply.isAttachment)

	reply = conv.replies[1]
	require.Equal(t, "Successfully checked in host for <@user>!", reply.text)
	f := messageField{"Host", ":desktop_computer: 1.2.3.4"}
	require.Equal(t, f, reply.fields[0])
	require.Equal(t, "good", reply.color)
	require.Empty(t, reply.timestamp, "expected reply 1 to be its own message")

	require.False(t, backend.(*DebugBackend).isCheckedOut)
}

func TestCheckInHostAlreadyIn(t *testing.T) {
	resetBackend()

	conv := newTestConversation("check in host")
	CheckInHost(context.TODO(), conv)

	reply := conv.replies[0]
	require.Equal(t, "Sorry, <@user>! Looks like there isn't a host checked out right now!", reply.text)
}

func TestBaseImages(t *testing.T) {
	resetBackend()

	conv := newTestConversation("base images")
	BaseImages(context.TODO(), conv)

	reply := conv.replies[0]
	require.Equal(t, "<@user>: \n• `debug-base-image-1`\n• `debug-base-image-2`\n• `debug-base-image-3`", reply.text)
}

func TestRestoreBackup(t *testing.T) {
	resetBackend()

	conv := newTestConversation("restore backup debug-base-image-2")
	props := proper.NewProperties(map[string]string{
		"image": "debug-base-image-2",
	})
	conv.SetProperties(props)

	RestoreBackup(context.TODO(), conv)

	require.Len(t, conv.replies, 2)

	reply := conv.replies[0]
	require.Equal(t, "Restoring backup for <@user>…", reply.text)
	require.True(t, reply.isAttachment)
	f := messageField{"Image", "debug-base-image-2"}
	require.Equal(t, f, reply.fields[0])

	reply = conv.replies[1]
	require.Equal(t, "Successfully restored backup for <@user>!", reply.text)
	require.Equal(t, f, reply.fields[0])
	require.Equal(t, "good", reply.color)
}
