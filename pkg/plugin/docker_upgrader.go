// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
package plugin

import (
	"context"
	"time"

	"go.blockdaemon.com/bpm/sdk/pkg/docker"
	"go.blockdaemon.com/bpm/sdk/pkg/node"
)

// DockerUpgrader provides a default strategy for upgrading docker based nodes
//
// The default upgrade strategy uses a LifecycleHandler to remove all containers. If they where running they get started again which will pull new container images.
//
// This works as long as only the container versions change. If the the upgrade needs changes to the configs or migrations tasks it is
// recommended to provide a custom Upgrader.
type DockerUpgrader struct {
	containers []docker.Container
}

// NewDockerUpgrader instantiates DockerUpgrader
func NewDockerUpgrader(containers []docker.Container) DockerUpgrader {
	return DockerUpgrader{containers: containers}
}

// Upgrade upgrades all containers by removing and starting them again
func (d DockerUpgrader) Upgrade(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Which containers are currently running?
	runningContainers := []docker.Container{}
	for _, container := range d.containers {
		running, err := client.IsContainerRunning(ctx, container.Name)
		if err != nil {
			return err
		}
		if running {
			runningContainers = append(runningContainers, container)
		}
	}

	// Remove containers
	for _, container := range d.containers {
		if err = client.ContainerAbsent(ctx, container); err != nil {
			return err
		}
	}

	// Start containers that where previously running (this will pull the new versions)
	for _, container := range runningContainers {
		if err = client.ContainerRuns(ctx, container); err != nil {
			return err
		}
	}

	return nil
}
