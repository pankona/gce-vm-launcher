package cloudfuncs

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/pankona/gce-vm-launcher/gce"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

func Command(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	arg, err := getArgument(q)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if _, err = w.Write([]byte(err.Error())); err != nil {
			log.Printf("failed to write response: %v", err)
		}
		return
	}

	switch arg {
	case "start":
		start(w)
	case "stop":
		stop(w)
	case "status":
		status(w)
	}
}

func getArgument(v url.Values) (string, error) {
	arg := v["arg"]
	if len(arg) == 0 {
		return "", fmt.Errorf("query parameter arg is missing")
	}
	switch arg[0] {
	case "start":
	case "stop":
	case "status":
	default:
		return "", fmt.Errorf("unsupported arg. Please specify one of start, stop or status.")
	}

	return arg[0], nil
}

type statusStore struct {
	projectID string
}

type Status struct {
	Time   time.Time `datastore:"time"`
	Status string    `datastore:"status"`
}

func (ss *statusStore) Save(ctx context.Context, status gce.GCEStatus) error {
	s := Status{
		Time:   status.Time,
		Status: status.Status,
	}

	client, err := datastore.NewClient(ctx, ss.projectID)
	if err != nil {
		return err
	}

	key := datastore.IncompleteKey("Status", nil)
	_, err = client.Put(ctx, key, &s)

	return err
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
		Project:     os.Getenv("GCE_VM_LAUNCHER_PROJECT"),
		Zone:        os.Getenv("GCE_VM_LAUNCHER_ZONE"),
		Instance:    os.Getenv("GCE_VM_LAUNCHER_INSTANCE"),
		StatusStore: &statusStore{projectID: os.Getenv("GCE_VM_LAUNCHER_PROJECT")},
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
		if _, err = w.Write([]byte("accepted: [start]")); err != nil {
			log.Printf("failed write response: %v", err)
		}
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
		if _, err = w.Write([]byte("accepted: [stop]")); err != nil {
			log.Printf("failed write response: %v", err)
		}
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
		if _, err = w.Write([]byte(fmt.Sprintf("status: %v, external ip: %v\n", status, externalIP))); err != nil {
			log.Printf("failed write response: %v", err)
		}
	})
}

func StoreStatus(w http.ResponseWriter, r *http.Request) {
	withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) {
		status, externalIP, err := g.GetStatus(ctx, computeService)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
			if err != nil {
				log.Printf("failed write response: %v", err)
			}
			return
		}

		err = g.WriteStatus(ctx, status)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
			if err != nil {
				log.Printf("failed write status to data repository: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write([]byte(fmt.Sprintf("status: %v, external ip: %v\n", status, externalIP))); err != nil {
			log.Printf("failed write response: %v", err)
		}
	})
}
