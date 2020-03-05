// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
package plugin

import (
	"bytes"
	"context"
	"io/ioutil"
	"path"
	"text/template"
	"time"

	"github.com/Blockdaemon/bpm-sdk/pkg/docker"
	"github.com/Blockdaemon/bpm-sdk/pkg/node"

	homedir "github.com/mitchellh/go-homedir"
)

// DockerLifecycleHandler provides functions to manage a node using plain docker containers
type DockerLifecycleHandler struct {
	containers []docker.Container
}

const (
	filebeatContainerImage = "docker.elastic.co/beats/filebeat:7.4.1"
	filebeatContainerName  = "filebeat"
	filebeatConfigFile     = "filebeat.yml"
	filebeatConfigTpl      = `filebeat.inputs:
- type: container
  paths: 
  - '/var/lib/docker/containers/*/*.log'
fields:
    node:
        launch_type: bpm
        xid: {{ .Node.ID }}
fields_under_root: true
output:
{{- if .Node.Collection.Host }}
    logstash:
        hosts:
        - "{{ .Node.Collection.Host }}"
        ssl:
            certificate: /etc/ssl/beats/beat.crt
            certificate_authorities:
            - /etc/ssl/beats/ca.crt
            key: /etc/ssl/beats/beat.key
{{- else }}
    console:
        pretty: true
{{- end }}
`
)

func NewDockerLifecycleHandler(containers []docker.Container) DockerLifecycleHandler {
	return DockerLifecycleHandler{containers: containers}
}

// Start starts monitoring agents and delegates to another function to start blockchain containers
func (d DockerLifecycleHandler) Start(currentNode node.Node) error {
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
	if currentNode.Collection.Host != "" {
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

		// Render filebeat config file
		outputFilename := path.Join(currentNode.NodeDirectory(), filebeatConfigFile)
		tmpl, err := template.New(outputFilename).Parse(filebeatConfigTpl)
		if err != nil {
			return err
		}
		output := bytes.NewBufferString("")
		err = tmpl.Execute(output, currentNode)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(outputFilename, output.Bytes(), 0644); err != nil {
			return err
		}

		// Start filebeat container
		filebeatContainer := docker.Container{
			Name:  filebeatContainerName,
			Image: filebeatContainerImage,
			Cmd:   []string{"-e", "-strict.perms=false"},
			// using the first containers network is a decent default, if we ever do mult-network deployments we may need to rethink this
			NetworkID: d.containers[0].NetworkID,
			Mounts: []docker.Mount{
				{
					Type: "bind",
					From: outputFilename,
					To:   "/usr/share/filebeat/filebeat.yml",
				},
				{
					Type: "bind",
					From: "/var/lib/docker/containers",
					To:   "/var/lib/docker/containers",
				},
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
			},
			User: "root",
		}

		if err := client.ContainerRuns(ctx, filebeatContainer); err != nil {
			return err
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
func (d DockerLifecycleHandler) Status(currentNode node.Node) (string, error) {
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

// DockerStop removes all containers
func (d DockerLifecycleHandler) Stop(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode.NamePrefix(), currentNode.ConfigsDirectory())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for _, container := range d.containers {
		if err = client.ContainerStopped(ctx, container.Name, container.NetworkID); err != nil {
			return err
		}
	}

	return nil
}

// Removes any data (typically the blockchain itself) related to the node
func (d DockerLifecycleHandler) RemoveData(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode.NamePrefix(), currentNode.ConfigsDirectory())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
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

	return nil
}

// Removes the docker network and containers
func (d DockerLifecycleHandler) RemoveRuntime(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode.NamePrefix(), currentNode.ConfigsDirectory())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	for _, container := range d.containers {
		if err = client.ContainerAbsent(ctx, container.Name, container.NetworkID); err != nil {
			return err
		}
	}

	// Remove network(s)
	for _, container := range d.containers {
		if err := client.NetworkAbsent(ctx, container.NetworkID); err != nil {
			return err
		}
	}

	return nil
}
