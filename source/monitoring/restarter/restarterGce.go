package restarter

import (
    "fmt"
    "log"

    "golang.org/x/net/context"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/compute/v1"
)

const (
	gceRegion = "us-central1-a"
)

type GceRestarter struct {
	projectID string
	client *compute.Service
	ctx context.Context
}

func NewGce(projectID string) *GceRestarter {
	ctx := context.Background()

	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
    if err != nil {
            log.Fatal(err)
    }
    computeService, err := compute.New(c)
    if err != nil {
            log.Fatal(err)
    }
    return &GceRestarter{projectID, computeService, ctx}
}

func (Restarter *GceRestarter) start(instanceID string) error {
	_, err := Restarter.client.Instances.Start(Restarter.projectID, gceRegion, instanceID).Context(Restarter.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to start instance: %v", err)
	}
	return nil
}

func (Restarter *GceRestarter) reset(instanceID string) error {
	_, err := Restarter.client.Instances.Reset(Restarter.projectID, gceRegion, instanceID).Context(Restarter.ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to reset instance: %v", err)
	}
	return nil
}

func (Restarter *GceRestarter) getState(instanceID string) (string, error) {
	resp, err := Restarter.client.Instances.Get(Restarter.projectID, gceRegion, instanceID).Context(Restarter.ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get instance state: %v", err)
	}
	return resp.Status, nil
}

func (Restarter *GceRestarter) Restart(instanceID string) (bool, error) {
	state, err := Restarter.getState(instanceID)
	if err != nil {
		return false, err
	}
	fmt.Println(state)
	switch state {
	case "RUNNING":
		err := Restarter.reset(instanceID)
		return true, err
	case "TERMINATED", "STOPPED":
		err := Restarter.start(instanceID)
		return true, err
	}
	return false, err
}