package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pankona/gce-vm-launcher/gce"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func main() {
	if ok := validateArguments(); !ok {
		os.Exit(1)
	}

	ctx := context.Background()

	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err := compute.NewService(ctx, option.WithHTTPClient(c))
	if err != nil {
		log.Fatal(err)
	}

	request := os.Args[1]

	g := gce.GCE{
		Project:  "sponge-is-dry",
		Zone:     "asia-northeast1-b",
		Instance: "mario",
	}

	switch request {
	case "start":
		fallthrough
	case "stop":
		err = g.DoOperation(ctx, computeService, request)
	case "status":
		err = g.ShowStatus(ctx, computeService)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func validateArguments() bool {
	if len(os.Args) == 1 {
		fmt.Println("Not enough argument. Please specify one of start, stop or status.")
		return false
	}

	request := os.Args[1]
	switch request {
	case "start":
	case "stop":
	case "status":
	default:
		fmt.Println("Unsupported argument. Please specify one of start, stop or status.")
		return false
	}

	return true
}
