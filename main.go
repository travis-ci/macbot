package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"net/url"
	"os"
	"os/signal"
	"runtime/pprof"
	"strings"
)

var vSphereURL *url.URL
var vSphereInsecure = true

const (
	prodClusterPath   = "/pod-1/host/MacPro_Pod_1"
	packerClusterPath = "/pod-1/host/packer_image_dev"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var rtm *slack.RTM

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		measureCPUUsage(*cpuprofile)
	}
	setupInterruptHandler()

	var err error
	vSphereURL, err = url.Parse(os.Getenv("VSPHERE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("SLACK_API_TOKEN")
	api := slack.New(token)
	logger := log.New(os.Stderr, "macbot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(true)

	rtm = api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		ctx := context.Background()
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			dispatchCommand(ctx, ev)
		default:
			// ignore
		}
	}
}

type contextKey string

var (
	contextKeyMessage = contextKey("message")
)

func dispatchCommand(ctx context.Context, msg *slack.MessageEvent) {
	ctx = context.WithValue(ctx, contextKeyMessage, msg)

	switch strings.TrimSpace(msg.Text) {
	case "checked out", "is checked out":
		IsHostCheckedOut(ctx, msg)
	case "checkout host", "check out host":
		CheckOutHost(ctx, msg)
	case "checkin host", "check in host":
		CheckInHost(ctx, msg)
	default:
		// ignore all other messages
	}
}

func currentMessage(ctx context.Context) *slack.MessageEvent {
	return ctx.Value(contextKeyMessage).(*slack.MessageEvent)
}

func reply(ctx context.Context, text string, args ...interface{}) {
	msg := currentMessage(ctx)
	text = fmt.Sprintf(text, args...)

	isDM := strings.HasPrefix(msg.Channel, "D")
	if !isDM {
		text = fmt.Sprintf("<@%s>: %s", msg.User, text)
	}

	m := rtm.NewOutgoingMessage(text, msg.Channel)
	rtm.SendMessage(m)
}

func typing(ctx context.Context) {
	msg := currentMessage(ctx)
	rtm.SendMessage(rtm.NewTypingMessage(msg.Channel))
}

func measureCPUUsage(profile string) {
	f, err := os.Create(profile)
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
}

func setupInterruptHandler() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		log.Println("received interrupt, shutting down...")
		if *cpuprofile != "" {
			log.Println("writing CPU profile...")
			pprof.StopCPUProfile()
		}
		os.Exit(0)
	}()
}
