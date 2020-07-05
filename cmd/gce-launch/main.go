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
		Project:  os.Getenv("GCE_VM_LAUNCHER_PROJECT"),
		Zone:     os.Getenv("GCE_VM_LAUNCHER_ZONE"),
		Instance: os.Getenv("GCE_VM_LAUNCHER_INSTANCE"),
	}

	switch request {
	case "start":
		fallthrough
	case "stop":
		err = g.DoOperation(ctx, computeService, request)
	case "status":
		err = showStatus(ctx, computeService, g)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func showStatus(ctx context.Context, computeService *compute.Service, g gce.GCE) error {
	status, externalIP, err := g.GetStatus(ctx, computeService)
	if err != nil {
		return err
	}

	log.Printf("status: %v, external ip: %v\n", status, externalIP)
	return nil
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

	if os.Getenv("GCE_VM_LAUNCHER_PROJECT") == "" {
		fmt.Println("Environment variable GCP_VM_LAUNCHER_PROJECT is not specified.")
		return false
	}
	if os.Getenv("GCE_VM_LAUNCHER_ZONE") == "" {
		fmt.Println("Environment variable GCP_VM_LAUNCHER_ZONE is not specified.")
		return false
	}
	if os.Getenv("GCE_VM_LAUNCHER_INSTANCE") == "" {
		fmt.Println("Environment variable GCP_VM_LAUNCHER_INSTANCE is not specified.")
		return false
	}

	return true
}
