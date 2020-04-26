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

func withComputeService(f func(ctx context.Context, computeService *compute.Service, g gce.GCE)) {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err := compute.New(c)
	if err != nil {
		log.Fatal(err)
	}

	g := gce.GCE{
		Project:  "sponge-is-dry",
		Zone:     "asia-northeast1-b",
		Instance: "mario",
	}

	f(ctx, computeService, g)
}

func Start(w http.ResponseWriter, r *http.Request) {
	withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) {
		err := g.DoOperation(ctx, computeService, "start")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
		}
		w.WriteHeader(http.StatusOK)
	})
}

func Stop(w http.ResponseWriter, r *http.Request) {
	withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) {
		err := g.DoOperation(ctx, computeService, "stop")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
		}
		w.WriteHeader(http.StatusOK)
	})
}

func Status(w http.ResponseWriter, r *http.Request) {
	withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) {
		status, externalIP, err := g.GetStatus(ctx, computeService)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("status: %v, external ip: %v\n", status, externalIP)))
	})
}
