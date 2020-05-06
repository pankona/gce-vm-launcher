package gce

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/api/compute/v1"
)

type GCE struct {
	Project  string
	Zone     string
	Instance string
}

func (g *GCE) DoOperation(ctx context.Context, computeService *compute.Service, arg string) error {
	var err error
	switch arg {
	case "start":
		log.Println("starting instance")
		if _, err = computeService.Instances.Start(g.Project, g.Zone, g.Instance).Context(ctx).Do(); err != nil {
			return err
		}
	case "stop":
		log.Println("stopping instance")
		if _, err = computeService.Instances.Stop(g.Project, g.Zone, g.Instance).Context(ctx).Do(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown argument [%s] is specified", arg)
	}

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
