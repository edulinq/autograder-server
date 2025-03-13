package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

func CanAccessDocker() bool {
	docker, err := getDockerClient()
	if docker != nil {
		defer docker.Close()
	}

	if err != nil {
		return false
	}

	_, err = docker.Info(context.Background())
	if err != nil {
		return false
	}

	return true
}

func getDockerClient() (*client.Client, error) {
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("Cannot create Docker client: '%w'.", err)
	}

	return docker, nil
}
