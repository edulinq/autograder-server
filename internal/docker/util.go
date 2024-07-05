package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

func CanAccessDocker() bool {
	_, docker, err := getDockerClient()
	if docker != nil {
		defer docker.Close()
	}

	return (err == nil)
}

func getDockerClient() (context.Context, *client.Client, error) {
	ctx := context.Background()
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return ctx, nil, fmt.Errorf("Cannot create Docker client: '%w'.", err)
	}

	return ctx, docker, nil
}
