package funcs

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/pankona/gce-vm-launcher/gce"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

func Start(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err := compute.New(c)
	if err != nil {
		log.Fatal(err)
	}

	err = gce.DoOperation(ctx, computeService, "start")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
	}
	w.WriteHeader(http.StatusOK)
}

func Stop(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err := compute.New(c)
	if err != nil {
		log.Fatal(err)
	}

	err = gce.DoOperation(ctx, computeService, "stop")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
	}
	w.WriteHeader(http.StatusOK)
}

func Status(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err := compute.New(c)
	if err != nil {
		log.Fatal(err)
	}

	status, externalIP, err := gce.GetStatus(ctx, computeService)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("status: %v, external ip: %v\n", status, externalIP)))
}
