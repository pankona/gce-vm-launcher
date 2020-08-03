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
	"google.golang.org/api/iterator"
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

func (ss *statusStore) lastRecord(ctx context.Context, client *datastore.Client) (*datastore.Key, Status, error) {
	q := datastore.NewQuery("Status").Order("-time").Limit(1)
	it := client.Run(ctx, q)

	var (
		s   Status
		ret *datastore.Key
	)

	for {
		key, err := it.Next(&s)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, Status{}, err
		}
		ret = key
	}

	return ret, s, nil
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
	key, st, err := ss.lastRecord(ctx, client)
	if err != nil {
		return err
	}

	if status.Status == "TERMINATED" && st.Status == "TERMINATED" {
		err = client.Delete(ctx, key)
		if err != nil {
			log.Printf("failed to delete continuous TERMINATED: %v\n", err)
		}
	}

	key = datastore.IncompleteKey("Status", nil)
	_, err = client.Put(ctx, key, &s)

	return err
}

func withComputeService(f func(ctx context.Context, computeService *compute.Service, g gce.GCE) error) error {
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		return err
	}

	computeService, err := compute.NewService(ctx, option.WithHTTPClient(c))
	if err != nil {
		return err
	}

	g := gce.GCE{
		Project:     os.Getenv("GCE_VM_LAUNCHER_PROJECT"),
		Zone:        os.Getenv("GCE_VM_LAUNCHER_ZONE"),
		Instance:    os.Getenv("GCE_VM_LAUNCHER_INSTANCE"),
		StatusStore: &statusStore{projectID: os.Getenv("GCE_VM_LAUNCHER_PROJECT")},
	}

	return f(ctx, computeService, g)
}

func start(w http.ResponseWriter) {
	_ = withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) error {
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

		return nil
	})
}

func stop(w http.ResponseWriter) {
	_ = withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) error {
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

		return nil
	})
}

func status(w http.ResponseWriter) {
	_ = withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) error {
		status, externalIP, err := g.GetStatus(ctx, computeService)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, err = w.Write([]byte(fmt.Sprintf("%s\n", err.Error())))
			if err != nil {
				return fmt.Errorf("failed write response: %v", err)
			}
		}
		w.WriteHeader(http.StatusOK)
		if _, err = w.Write([]byte(fmt.Sprintf("status: %v, external ip: %v\n", status, externalIP))); err != nil {
			return fmt.Errorf("failed write response: %v", err)
		}

		return nil
	})
}

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func StoreStatus(ctx context.Context, _ PubSubMessage) error {
	return withComputeService(func(ctx context.Context, computeService *compute.Service, g gce.GCE) error {
		status, _, err := g.GetStatus(ctx, computeService)
		if err != nil {
			return err
		}

		err = g.WriteStatus(ctx, status)
		if err != nil {
			return err
		}

		return nil
	})
}
