// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
package plugin

import (
	"context"

	"fmt"
	"strings"
	"time"

	"github.com/Blockdaemon/bpm-sdk/pkg/docker"
	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/Blockdaemon/bpm-sdk/pkg/template"

	homedir "github.com/mitchellh/go-homedir"
)

// DockerPlugin is a varation of Plugin that comes with default methods for docker
type DockerPlugin struct {
	// All configurations files and their respective templates
	configFilesAndTemplates map[string]string

	// All containers that should be managed by this plugin
	containers []docker.Container

	name        string
	version     string
	description string
}

const (
	filebeatContainerImage = "docker.elastic.co/beats/filebeat:7.3.1"
	filebeatContainerName  = "filebeat"
	filebeatConfigFile     = "filebeat.yml"
)

func NewDockerPlugin(name, description, version string, containers []docker.Container, configFilesAndTemplates map[string]string) Plugin {
	// Add filebeat to the passed in containers
	filebeatContainer := docker.Container{
		Name:      filebeatContainerName,
		Image:     filebeatContainerImage,
		Cmd:       []string{"-e", "-strict.perms=false"},
		// using the first containers network is a decent default, if we ever do mult-network deployments we may need to rethink this
		NetworkID: containers[0].NetworkID,
		Mounts: []docker.Mount{
			{
				Type: "bind",
				From: filebeatConfigFile,
				To:   "/usr/share/filebeat/filebeat.yml",
			},
		},
		User: "root",
	}

	containers = append(containers, filebeatContainer)
	configFilesAndTemplates[filebeatConfigFile] = filebeatConfigTpl

	return DockerPlugin{
		configFilesAndTemplates: configFilesAndTemplates,
		containers:              containers,
		name:                    name,
		description:             description,
		version:                 version,
	}
}

func (d DockerPlugin) Name() string {
	return d.name
}

func (d DockerPlugin) Version() string {
	return d.version
}

func (d DockerPlugin) Description() string {
	return d.description
}

// CreateSecrets does nothing except printing that it does nothing
func (d DockerPlugin) CreateSecrets(currentNode node.Node) error {
	fmt.Println("Nothing to do here, skipping create-secrets")
	return nil
}

// Upgrade does nothing except printing that it does nothing
func (d DockerPlugin) Upgrade(currentNode node.Node) error {
	fmt.Println("Nothing to do here, skipping upgrade")
	return nil
}

// Test does nothing except printing that it does nothing
func (d DockerPlugin) Test(currentNode node.Node) (bool, error) {
	fmt.Println("Nothing to do here, skipping test")
	return true, nil
}

// CreateConfigs creates configuration files for the blockchain client
func (d DockerPlugin) CreateConfigs(currentNode node.Node) error {
	return template.ConfigFilesRendered(d.configFilesAndTemplates, template.TemplateData{
		Node: currentNode,
		Data: map[string]interface{}{
			"Containers": d.containers,
		},
	})
}

// Start starts monitoring agents and delegates to another function to start blockchain containers
func (d DockerPlugin) Start(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode.NamePrefix(), currentNode.ConfigsDirectory())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// First, create the docker network(s) if they don't exist yet
	for _, container := range d.containers {
		if err := client.NetworkExists(ctx, container.NetworkID); err != nil {
			return err
		}
	}

	//////////////////////////////////////////////////////////////////////////////
	// TODO: This is just temporarily until we have a proper authentication system
	//////////////////////////////////////////////////////////////////////////////
	currentNode.Collection.Key, err = homedir.Expand(currentNode.Collection.Key)
	if err != nil {
		return err
	}
	currentNode.Collection.Cert, err = homedir.Expand(currentNode.Collection.Cert)
	if err != nil {
		return err
	}
	currentNode.Collection.CA, err = homedir.Expand(currentNode.Collection.CA)
	if err != nil {
		return err
	}

	for ix, container := range d.containers {
		if strings.HasSuffix(container.Name, "beat") { // yeah, I know, super hacky but it's just temporarily
			fmt.Printf("Add ssl certs to container: %s\n", container.Name)
			sslMounts := []docker.Mount{
				{
					Type: "bind",
					From: currentNode.Collection.CA,
					To:   "/etc/ssl/beats/ca.crt",
				},
				{
					Type: "bind",
					From: currentNode.Collection.Cert,
					To:   "/etc/ssl/beats/beat.crt",
				},
				{
					Type: "bind",
					From: currentNode.Collection.Key,
					To:   "/etc/ssl/beats/beat.key",
				},
			}
			d.containers[ix].Mounts = append(container.Mounts, sslMounts...)
		}
	}
	//////////////////////////////////////////////////////////////////////////////
	// TODO end
	//////////////////////////////////////////////////////////////////////////////

	// Next, start the containers
	for _, container := range d.containers {
		if err := client.ContainerRuns(ctx, container); err != nil {
			return err
		}
	}

	return nil
}

// DockerStatus returns the status of the running blockchain client and monitoring containers
func (d DockerPlugin) Status(currentNode node.Node) (string, error) {
	client, err := docker.NewBasicManager(currentNode.NamePrefix(), currentNode.ConfigsDirectory())
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	containersRunning := 0

	for _, container := range d.containers {
		running, err := client.IsContainerRunning(ctx, container.Name)
		if err != nil {
			return "", err
		}
		if running {
			containersRunning += 1
		}
	}

	if containersRunning == 0 {
		return "stopped", nil
	} else if len(d.containers) == containersRunning {
		return "running", nil
	}

	return "incomplete", nil
}

// DockerStop removes all configuration files and containers, volumes, network based on naming conventions
//
// Container names and volume names for a particular node all start with "bd-<node-id>".
func (d DockerPlugin) Stop(currentNode node.Node, purge bool) error {
	client, err := docker.NewBasicManager(currentNode.NamePrefix(), currentNode.ConfigsDirectory())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for _, container := range d.containers {
		if err = client.ContainerAbsent(ctx, container.Name); err != nil {
			return err
		}
	}

	// Remove network(s)
	for _, container := range d.containers {
		if err := client.NetworkAbsent(ctx, container.NetworkID); err != nil {
			return err
		}
	}

	if purge {
		// Remove volumes
		for _, container := range d.containers {
			for _, mount := range container.Mounts {
				if mount.Type == "volume" {
					if err = client.VolumeAbsent(ctx, mount.From); err != nil {
						return err
					}
				}
			}
		}

		// Remove all configuration files
		for file := range d.configFilesAndTemplates {
			if err := template.ConfigFileAbsent(file, currentNode); err != nil {
				return err
			}
		}
	}

	return nil
}
