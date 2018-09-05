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
	if !strings.HasSuffix(reply, "There is no host checked out for building images.") {
		t.Fatal("unexpected reply, got", reply)
	}

	// now check out a host and try again
	host, _ := backend.SelectHost(context.TODO())
	backend.CheckOutHost(context.TODO(), host)

	conv = newTestConversation("is checked out")
	IsHostCheckedOut(context.TODO(), conv)

	reply = conv.replies[0]
	if !strings.HasSuffix(reply, "There is a host currently checked out for building images.") {
		t.Fatal("unexpected reply, got", reply)
	}
}
