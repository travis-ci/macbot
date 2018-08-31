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

var backend Backend

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var debug = flag.Bool("debug", false, "use debugging backend, don't talk to vsphere")

var rtm *slack.RTM

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		measureCPUUsage(*cpuprofile)
	}
	setupInterruptHandler()

	setupBackend()

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

	switch messageCommand(msg) {
	case "checked out", "is checked out":
		go IsHostCheckedOut(ctx, msg)
	case "checkout host", "check out host":
		go CheckOutHost(ctx, msg)
	case "checkin host", "check in host":
		go CheckInHost(ctx, msg)
	default:
		// ignore all other messages
	}
}

func messageCommand(msg *slack.MessageEvent) string {
	userID := rtm.GetInfo().User.ID
	text := strings.TrimSpace(msg.Text)

	mentionPrefix := "<@" + userID + "> "
	if !isDirectMessage(msg) && !strings.HasPrefix(text, mentionPrefix) {
		return ""
	}

	if strings.HasPrefix(text, mentionPrefix) {
		text = text[len(mentionPrefix):len(text)]
	}

	return strings.TrimSpace(text)
}

func currentMessage(ctx context.Context) *slack.MessageEvent {
	return ctx.Value(contextKeyMessage).(*slack.MessageEvent)
}

func isDirectMessage(msg *slack.MessageEvent) bool {
	return strings.HasPrefix(msg.Channel, "D")
}

func reply(ctx context.Context, text string, args ...interface{}) {
	msg := currentMessage(ctx)
	text = fmt.Sprintf(text, args...)

	if !isDirectMessage(msg) {
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

func setupBackend() {
	if *debug {
		setupDebugBackend()
	} else {
		setupVSphereBackend()
	}
}

func setupDebugBackend() {
	backend = &DebugBackend{
		Host: DebugHost("1.2.3.4"),
	}
}

func setupVSphereBackend() {
	vSphereURL, err := url.Parse(os.Getenv("VSPHERE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	backend = &VSphereBackend{
		URL:             vSphereURL,
		Insecure:        true,
		ProdClusterPath: "/pod-1/host/MacPro_Pod_1",
		DevClusterPath:  "/pod-1/host/packer-image-dev",
	}
}
