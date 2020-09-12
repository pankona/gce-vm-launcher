package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

func main() {
	projectID := os.Getenv("GCE_VM_LAUNCHER_PROJECT")
	if projectID == "" {
		fmt.Println("error: please specify GCE_VM_LAUNCHER_PROJECT environment variable.")
		os.Exit(1)
	}
	Run(context.Background(), projectID)
}

func Run(ctx context.Context, projectID string) {
	month := flag.String("month", time.Now().Format("2006-01"), "specify year and month to get vm uptime (yyyy-mm)")
	flag.Parse()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	ss, err := monthlyRecords(ctx, client, *month)
	if err != nil {
		panic(err)
	}

	var runningCount int
	for _, s := range ss {
		if s.Status == "RUNNING" {
			runningCount += 1
		}
	}
	fmt.Printf("running time: %d minutes (%dh %dm)", runningCount, runningCount/60, runningCount%60)
}

type Status struct {
	Time   time.Time `datastore:"time"`
	Status string    `datastore:"status"`
}

func monthlyRecords(ctx context.Context, client *datastore.Client, month string) ([]Status, error) {
	from, err := time.Parse("2006-01", month)
	if err != nil {
		panic(err)
	}
	to := endOfMonth(from)

	fmt.Printf("calculating uptime of %v\n", from.Format("2006-01"))

	q := datastore.NewQuery("Status").
		Filter("time >=", from).
		Filter("time <=", to).
		Order("time")
	it := client.Run(ctx, q)

	var (
		ret []Status
		s   Status
	)
	for {
		_, err := it.Next(&s)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		ret = append(ret, s)
	}

	return ret, nil
}

func endOfMonth(from time.Time) time.Time {
	y, m, _ := from.Date()
	loc := from.Location()
	nextMonth := time.Date(y, m+1, 1, 0, 0, 0, 0, loc)
	return nextMonth.Add(-1 * time.Second)
}
