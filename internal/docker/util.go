package docker

import (
	"fmt"

	"github.com/docker/docker/client"
)

func CanAccessDocker() bool {
	docker, err := getDockerClient()
	if docker != nil {
		defer docker.Close()
	}

	return (err == nil)
}

func getDockerClient() (*client.Client, error) {
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("Cannot create Docker client: '%w'.", err)
	}

	return docker, nil
}
