package cloudfuncs

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/pankona/gce-vm-launcher/gce"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func Launch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	err := validateArguments(q)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	request := q["req"][0]
	switch request {
	case "start":
		start(w)
	case "stop":
		stop(w)
	case "status":
		status(w)
	}
}

func validateArguments(v url.Values) error {
	request := v["req"]
	if len(request) == 0 {
		return fmt.Errorf("query parameter req is missing")
	}
	switch request[0] {
	case "start":
	case "stop":
	case "status":
	default:
		return fmt.Errorf("unsupported req. Please specify one of start, stop or status.")
	}

	return nil
}

func withComputeService(f func(ctx context.Context, computeService *compute.Service, g gce.GCE)) {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err := compute.NewService(ctx, option.WithHTTPClient(c))
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

func start(w http.ResponseWriter) {
	withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) {
		err := g.DoOperation(ctx, computeService, "start")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
			if err != nil {
				log.Printf("failed write response: %v", err)
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}

func stop(w http.ResponseWriter) {
	withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) {
		err := g.DoOperation(ctx, computeService, "stop")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
			if err != nil {
				log.Printf("failed write response: %v", err)
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}

func status(w http.ResponseWriter) {
	withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) {
		status, externalIP, err := g.GetStatus(ctx, computeService)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
			if err != nil {
				log.Printf("failed write response: %v", err)
			}
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(fmt.Sprintf("status: %v, external ip: %v\n", status, externalIP)))
		if err != nil {
			log.Printf("failed write response: %v", err)
		}
	})
}
