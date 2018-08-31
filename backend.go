package main

import (
	"context"
	"github.com/travis-ci/vsphere-images"
	"github.com/vmware/govmomi/object"
	"net/url"
	"time"
)

type Host interface {
	Name() string
}

type Backend interface {
	IsHostCheckedOut(context.Context) (bool, error)
	SelectHost(context.Context) (Host, error)
	CheckOutHost(context.Context, Host) error
	CheckInHost(context.Context) (Host, error)
}

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

type DebugHost string

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
