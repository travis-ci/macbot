package main

import (
	"context"
	"flag"
	"github.com/nlopes/slack"
	"log"
	"net/url"
	"os"
	"os/signal"
	"runtime/pprof"
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

	router := NewRouter()
	router.HandleFunc("base images", BaseImages)
	router.HandleFunc("base vms", BaseImages)
	router.HandleFunc("restore backup <image>", RestoreBackup)
	router.HandleFunc("checked out", IsHostCheckedOut)
	router.HandleFunc("is checked out", IsHostCheckedOut)
	router.HandleFunc("checkout host", CheckOutHost)
	router.HandleFunc("check out host", CheckOutHost)
	router.HandleFunc("checkin host", CheckInHost)
	router.HandleFunc("check in host", CheckInHost)

	for msg := range rtm.IncomingEvents {
		ctx := context.Background()
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			dispatchCommand(ctx, router, ev)
		default:
			// ignore
		}
	}
}

func dispatchCommand(ctx context.Context, router *Router, msg *slack.MessageEvent) {
	conv := NewConversation(msg)
	go router.Reply(ctx, conv)
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
	pod1URL, err := url.Parse(os.Getenv("VSPHERE_POD1_URL"))
	if err != nil {
		log.Fatal(err)
	}

	pod2URL, err := url.Parse(os.Getenv("VSPHERE_POD2_URL"))
	if err != nil {
		log.Fatal(err)
	}

	backend = &VSphereBackend{
		Pod1: DatacenterConfig{
			URL:             pod1URL,
			Insecure:        true,
			ProdClusterPath: "/pod-1/host/MacPro_Pod_1",
			DevClusterPath:  "/pod-1/host/packer_image_dev",
			BaseImagePath:   "/pod-1/vm/Base VMs",
		},
		Pod2: DatacenterConfig{
			URL:             pod2URL,
			Insecure:        true,
			ProdClusterPath: "/pod-2/host/MacPro_Pod_2",
			BaseImagePath:   "/pod-2/vm/Base VMs",
			BackupImagePath: "/pod-2/vm/VM Backups",
		},
	}
}
