package main

import (
	"context"
	"strings"
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
	if !strings.HasSuffix(reply.text, "There is no host checked out for building images.") {
		t.Fatal("unexpected reply, got", reply)
	}

	// now check out a host and try again
	host, _ := backend.SelectHost(context.TODO())
	backend.CheckOutHost(context.TODO(), host)

	conv = newTestConversation("is checked out")
	IsHostCheckedOut(context.TODO(), conv)

	reply = conv.replies[0]
	if !strings.HasSuffix(reply.text, "There is a host currently checked out for building images.") {
		t.Fatal("unexpected reply, got", reply)
	}
}

func TestCheckOutHost(t *testing.T) {
	resetBackend()
	conv := newTestConversation("check out host")
	CheckOutHost(context.TODO(), conv)

	if len(conv.replies) != 3 {
		t.Fatal("expected 3 replies, got", len(conv.replies))
	}
	reply := conv.replies[0]
	if reply.text != "Choosing a host to check out for <@user>…" {
		t.Fatal("unexpected reply 0, got", reply.text)
	}
	if !reply.isAttachment {
		t.Fatal("expected reply 0 to be an attachment")
	}

	reply = conv.replies[1]
	if reply.text != "Checking out host for <@user>…" {
		t.Fatal("unexpected reply 1, got", reply.text)
	}
	f := messageField{"Host", ":desktop_computer: 1.2.3.4"}
	if reply.fields[0] != f {
		t.Fatal("expected host field for fake host, got", reply.fields[0])
	}
	if reply.timestamp == "" {
		t.Fatal("expected reply 1 to be updating reply 0")
	}

	reply = conv.replies[2]
	if reply.text != "Successfully checked out host for <@user>!" {
		t.Fatal("unexpected reply 2, got", reply.text)
	}
	if reply.fields[0] != f {
		t.Fatal("expected host field for fake host, got", reply.fields[0])
	}
	if reply.color != "good" {
		t.Fatal("expected reply 2 to be good color, got", reply.color)
	}
	if reply.timestamp != "" {
		t.Fatal("expected reply 2 to be its own message")
	}

	if !backend.(*DebugBackend).isCheckedOut {
		t.Fatal("expected fake host to be checked out")
	}
}

func TestCheckOutHostAlreadyOut(t *testing.T) {
	resetBackend()
	host, _ := backend.SelectHost(context.TODO())
	backend.CheckOutHost(context.TODO(), host)

	conv := newTestConversation("check out host")
	CheckOutHost(context.TODO(), conv)

	reply := conv.replies[0]
	if reply.text != "Sorry, <@user>! Looks like there's already a host checked out for building images!" {
		t.Fatal("unexpected reply, got", reply.text)
	}
}

func TestCheckInHost(t *testing.T) {
	resetBackend()
	host, _ := backend.SelectHost(context.TODO())
	backend.CheckOutHost(context.TODO(), host)

	conv := newTestConversation("check in host")
	CheckInHost(context.TODO(), conv)

	if len(conv.replies) != 2 {
		t.Fatal("expected 2 replies, got", len(conv.replies))
	}
	reply := conv.replies[0]
	if reply.text != "Checking the host in for <@user>…" {
		t.Fatal("unexpected reply 0, got", reply.text)
	}
	if !reply.isAttachment {
		t.Fatal("expected reply 0 to be an attachment")
	}

	reply = conv.replies[1]
	if reply.text != "Successfully checked in host for <@user>!" {
		t.Fatal("unexpected reply 1, got", reply.text)
	}
	f := messageField{"Host", ":desktop_computer: 1.2.3.4"}
	if reply.fields[0] != f {
		t.Fatal("expected host field for fake host, got", reply.fields[0])
	}
	if reply.color != "good" {
		t.Fatal("expected reply 1 to be good color, got", reply.color)
	}
	if reply.timestamp != "" {
		t.Fatal("expected reply 1 to be its own message")
	}

	if backend.(*DebugBackend).isCheckedOut {
		t.Fatal("expected fake host to be checked in")
	}
}

func TestCheckInHostAlreadyIn(t *testing.T) {
	resetBackend()

	conv := newTestConversation("check in host")
	CheckInHost(context.TODO(), conv)

	reply := conv.replies[0]
	if reply.text != "Sorry, <@user>! Looks like there isn't a host checked out right now!" {
		t.Fatal("unexpected reply, got", reply.text)
	}
}
