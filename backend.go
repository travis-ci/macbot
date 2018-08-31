package main

import (
	"context"
	"github.com/travis-ci/vsphere-images"
	"github.com/vmware/govmomi/object"
	"net/url"
	"time"
)

// Host represents a host machine that can be checked in or out.
type Host interface {
	// Name returns the name of the host, for display in chat messages.
	Name() string
}

// Backend is a common interface for operations the bot would perform against vSphere.
//
// The Backend interface simplifies the chat command logic and allows us to substitute in
// a debug backend for testing chat interactions.
type Backend interface {
	IsHostCheckedOut(context.Context) (bool, error)
	SelectHost(context.Context) (Host, error)
	CheckOutHost(context.Context, Host) error
	CheckInHost(context.Context) (Host, error)
}

// VSphereBackend is the default backend, which communicates with a vSphere instance.
type VSphereBackend struct {
	URL             *url.URL
	Insecure        bool
	ProdClusterPath string
	DevClusterPath  string
}

func (b *VSphereBackend) IsHostCheckedOut(ctx context.Context) (bool, error) {
	return vsphereimages.IsHostCheckedOut(ctx, b.URL, b.Insecure, b.DevClusterPath)
}

func (b *VSphereBackend) SelectHost(ctx context.Context) (Host, error) {
	return vsphereimages.SelectAvailableHost(ctx, b.URL, b.Insecure, b.ProdClusterPath)
}

func (b *VSphereBackend) CheckOutHost(ctx context.Context, h Host) error {
	return vsphereimages.CheckOutSelectedHost(ctx, b.URL, b.Insecure, h.(*object.HostSystem), b.DevClusterPath, newProgressLogger())
}

func (b *VSphereBackend) CheckInHost(ctx context.Context) (Host, error) {
	return vsphereimages.CheckInHost(ctx, b.URL, b.Insecure, b.DevClusterPath, b.ProdClusterPath, newProgressLogger())
}

// DebugHost is a host in the debug backend.
//
// It is just a wrapper around a string, so that it can implement the Host interface.
type DebugHost string

// DebugBackend is a fake backend that can be used to test the Slack bot without interacting
// with real hosts.
//
// A DebugBackend conceptually has a single fake host that is not checked out at process
// start. Selecting a host always selects this host, and checking it in or out always succeeds.
type DebugBackend struct {
	Host         DebugHost
	isCheckedOut bool
}

func (b *DebugBackend) IsHostCheckedOut(ctx context.Context) (bool, error) {
	return b.isCheckedOut, nil
}

func (b *DebugBackend) SelectHost(ctx context.Context) (Host, error) {
	time.Sleep(time.Second)
	return b.Host, nil
}

func (b *DebugBackend) CheckOutHost(ctx context.Context, h Host) error {
	time.Sleep(10 * time.Second)
	b.isCheckedOut = true
	return nil
}

func (b *DebugBackend) CheckInHost(ctx context.Context) (Host, error) {
	time.Sleep(time.Second)
	b.isCheckedOut = false
	return b.Host, nil
}

func (h DebugHost) Name() string {
	return string(h)
}
