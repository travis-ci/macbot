package main

import (
	"context"
	"github.com/sbstjn/hanu"
	"github.com/travis-ci/vsphere-images"
	"log"
	"net/url"
	"os"
)

var vSphereURL *url.URL
var vSphereInsecure bool

const (
	prodClusterPath   = "/pod-1/host/MacPro_Pod_1"
	packerClusterPath = "/pod-1/host/packer_image_dev"
)

func main() {
	var err error
	vSphereInsecure = true
	vSphereURL, err = url.Parse(os.Getenv("VSPHERE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("SLACK_API_TOKEN")
	slack, err := hanu.New(token)
	if err != nil {
		log.Fatal(err)
	}

	slack.Command("checked out", IsHostCheckedOut)
	slack.Command("is checked out", IsHostCheckedOut)
	slack.Command("checkout host", CheckOutHost)
	slack.Command("check out host", CheckOutHost)
	slack.Command("checkin host", CheckInHost)
	slack.Command("check in host", CheckInHost)

	log.Println("Starting Slack bot.")
	slack.Listen()
}

func IsHostCheckedOut(conv hanu.ConversationInterface) {
	isCheckedOut, err := vsphereimages.IsHostCheckedOut(context.TODO(), vSphereURL, vSphereInsecure, packerClusterPath)
	if err != nil {
		conv.Reply(":exclamation: Oops! I couldn't determine if a host is checked out already. `%s`", err)
		return
	}

	if isCheckedOut {
		conv.Reply(":white_check_mark: There is a host currently checked out for building images.")
	} else {
		conv.Reply(":heavy_multiplication_x: There is no host checked out for building images.")
	}
}

func CheckOutHost(conv hanu.ConversationInterface) {
	isCheckedOut, err := vsphereimages.IsHostCheckedOut(context.TODO(), vSphereURL, vSphereInsecure, packerClusterPath)
	if err != nil {
		conv.Reply(":exclamation: Oops! I couldn't determine if a host is currently checked out. `%s`", err)
		return
	}

	if isCheckedOut {
		conv.Reply("Looks like there's already a host checked out for building images!")
		return
	}

	conv.Reply("Checking out a host for image building. I'll let you know when it's ready!")
	host, err := vsphereimages.CheckOutHost(context.TODO(), vSphereURL, vSphereInsecure, prodClusterPath, packerClusterPath, newProgressLogger())
	if err != nil {
		conv.Reply(":exclamation: Oops! I couldn't check out the host. `%s`", err)
		return
	}

	conv.Reply(":white_check_mark: Done! Host %s is checked out for image building.", host.Name())
}

func CheckInHost(conv hanu.ConversationInterface) {
	isCheckedOut, err := vsphereimages.IsHostCheckedOut(context.TODO(), vSphereURL, vSphereInsecure, packerClusterPath)
	if err != nil {
		conv.Reply(":exclamation: Oops! I couldn't determine if a host is currently checked out. `%s`", err)
		return
	}

	if !isCheckedOut {
		conv.Reply("Looks like there isn't a host checked out right now!")
		return
	}

	conv.Reply("Checking the host back into the production cluster...")
	host, err := vsphereimages.CheckInHost(context.TODO(), vSphereURL, vSphereInsecure, packerClusterPath, prodClusterPath, newProgressLogger())
	if err != nil {
		conv.Reply(":exclamation: Oops! I couldn't check the host back in. `%s`", err)
		return
	}

	conv.Reply(":white_check_mark: Done! Host %s is back in production.", host.Name())
}
