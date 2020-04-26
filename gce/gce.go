package gce

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/compute/v1"
)

type GCE struct {
	Project  string
	Zone     string
	Instance string
}

func (g *GCE) DoOperation(ctx context.Context, computeService *compute.Service, request string) error {
	var err error
	if request == "start" {
		fmt.Println("starting instance")
		_, err = computeService.Instances.Start(g.Project, g.Zone, g.Instance).Context(ctx).Do()
		if err != nil {
			return err
		}
		for {
			status, _, err := g.GetStatus(ctx, computeService)
			if err != nil {
				return err
			}

			fmt.Printf("current status: %s\n", status)

			switch status {
			case "RUNNING":
				return nil
			}
			<-time.After(1 * time.Second)
		}
	} else if request == "stop" {
		fmt.Println("stopping instance")
		_, err = computeService.Instances.Stop(g.Project, g.Zone, g.Instance).Context(ctx).Do()
		if err != nil {
			return err
		}
		for {
			status, _, err := g.GetStatus(ctx, computeService)
			if err != nil {
				return err
			}

			fmt.Printf("current status: %s\n", status)

			switch status {
			case "TERMINATED":
				return nil
			}
			<-time.After(1 * time.Second)
		}
	}

	return fmt.Errorf("unknown request [%s] is specified", request)
}

func (g *GCE) ShowStatus(ctx context.Context, computeService *compute.Service) error {
	status, externalIP, err := g.GetStatus(ctx, computeService)
	if err != nil {
		return err
	}
	fmt.Printf("status: %v, external ip: %v\n", status, externalIP)
	return nil
}

func (g *GCE) GetStatus(ctx context.Context, computeService *compute.Service) (string, string, error) {
	resp, err := computeService.Instances.Get(g.Project, g.Zone, g.Instance).Context(ctx).Do()
	if err != nil {
		return "", "", err
	}

	var externalIP string
	for _, v := range resp.NetworkInterfaces {
		for _, v2 := range v.AccessConfigs {
			externalIP = v2.NatIP
		}
	}
	return resp.Status, externalIP, nil
}
