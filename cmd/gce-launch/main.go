package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/pankona/gce-vm-launcher/gce"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

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
		Project:     os.Getenv("GCE_VM_LAUNCHER_PROJECT"),
		Zone:        os.Getenv("GCE_VM_LAUNCHER_ZONE"),
		Instance:    os.Getenv("GCE_VM_LAUNCHER_INSTANCE"),
		StatusStore: &statusStore{projectID: os.Getenv("GCE_VM_LAUNCHER_PROJECT")},
	}

	switch request {
	case "start":
		fallthrough
	case "stop":
		err = g.DoOperation(ctx, computeService, request)
	case "status":
		err = showStatus(ctx, computeService, g)
	case "store-status":
		err = storeStatus(ctx, computeService, g)
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

func storeStatus(ctx context.Context, computeService *compute.Service, g gce.GCE) error {
	status, _, err := g.GetStatus(ctx, computeService)
	if err != nil {
		return err
	}

	return g.WriteStatus(ctx, status)
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
	case "store-status":
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
