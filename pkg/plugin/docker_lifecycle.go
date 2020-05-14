// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"go.blockdaemon.com/bpm/sdk/pkg/docker"
	"go.blockdaemon.com/bpm/sdk/pkg/fileutil"
	"go.blockdaemon.com/bpm/sdk/pkg/node"
	sdktemplate "go.blockdaemon.com/bpm/sdk/pkg/template"
)

// DockerLifecycleHandler provides functions to manage a node using plain docker containers
type DockerLifecycleHandler struct {
	containers []docker.Container
}

const (
	// LogsDirectory is the subdirectory under the node directory where logs are saved
	LogsDirectory          = "logs"
	filebeatContainerImage = "docker.elastic.co/beats/filebeat:7.4.1"
	filebeatContainerName  = "filebeat"
	filebeatConfigFile     = "filebeat.yml"
	filebeatBaseConfigTpl  = `filebeat.inputs:
- type: container
  paths:
  - '/var/lib/docker/containers/*/*.log'
fields:
  node:
    project: development
    protocol_type: {{ .Node.PluginName | ToUpper }}
    user_id: bpm
    xid: {{ .Node.ID }}
fields_under_root: true
processors:
- add_docker_metadata: null
{{- if .PluginData.Containers }}
- else.add_fields:
    fields.log_type: system
    target: ''
  if.or:
  {{- range $container := .PluginData.Containers }}
    {{- if $container.CollectLogs }}
  - equals.container.name: {{ $.Node.NamePrefix }}{{ $container.Name }}
    {{- end }}
  {{- end }}
  then.add_fields:
    fields.log_type: user
    target: ''
{{- end }}
- drop_event.when.not.equals.log_type: user
`
	filebeatConsoleConfigTpl = `output:
  console:
    pretty: true
`
)

// NewDockerLifecycleHandler creates an instance of DockerLifecycleHandler
func NewDockerLifecycleHandler(containers []docker.Container) DockerLifecycleHandler {
	return DockerLifecycleHandler{containers: containers}
}

// renderMonitoringConfig renders the configuration file for filebeat
//
// We can run either with monitoring forwarding enabled or disabled:
//
// - If disabled we just use the base config and add a console output to it
// - If enabled (via --monitoring-pack) we extract the monitoring pack which contains a filebeat output and combine it with the base config
func (d DockerLifecycleHandler) renderMonitoringConfig(monitoringPath string, currentNode node.Node) error {
	filebeatConfigTpl := ""

	if currentNode.StrParameters["monitoring-pack"] == "" {
		fmt.Println("Forwarding of monitoring is disabled. Specify `--monitoring-pack` to enable it.")
		// Instead of forwarding we'll just create filebeat with a simple log output
		filebeatConfigTpl = filebeatBaseConfigTpl + "\n" + filebeatConsoleConfigTpl
	} else {
		fmt.Println("Enabling forwarding of monitoring data.")

		if err := fileutil.ExtractTarGz(currentNode.StrParameters["monitoring-pack"], monitoringPath); err != nil {
			return err
		}

		monitoringPackConfig, err := ioutil.ReadFile(filepath.Join(monitoringPath, "config.tpl"))
		if err != nil {
			return err
		}
		filebeatConfigTpl = filebeatBaseConfigTpl + "\n" + string(monitoringPackConfig)
	}

	// Render filebeat config
	outputFilename := path.Join(currentNode.NodeDirectory(), "monitoring", filebeatConfigFile)
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}
	tmpl, err := template.New(outputFilename).Funcs(funcMap).Parse(filebeatConfigTpl)
	if err != nil {
		return err
	}
	templateData := sdktemplate.TemplateData{
		Node:       currentNode,
		PluginData: map[string]interface{}{"Containers": d.containers},
	}
	output := bytes.NewBufferString("")
	err = tmpl.Execute(output, templateData)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(outputFilename, output.Bytes(), 0644)
}

// SetUpEnvironment configures the monitoring agents
func (d DockerLifecycleHandler) SetUpEnvironment(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode)
	if err != nil {
		return err
	}

	// Create logs directory if it doesn't exist yet
	_, err = fileutil.MakeDirectory(currentNode.NodeDirectory(), LogsDirectory)
	if err != nil {
		return err
	}

	// Create data directory if it doesn't exist yet
	_, err = fileutil.MakeDirectory(client.AddBasePath(currentNode.StrParameters["data-dir"]))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Create the docker network if it doesn't exist yet
	if err := client.NetworkExists(ctx, currentNode.StrParameters["docker-network"]); err != nil {
		return err
	}

	// Create monitoring directory
	monitoringPath := client.AddBasePath("monitoring")
	_, err = fileutil.MakeDirectory(monitoringPath)
	if err != nil {
		return err
	}

	// Render the config
	return d.renderMonitoringConfig(monitoringPath, currentNode)
}

// TearDownEnvironment is currently just a placeholder that does nothing
func (d DockerLifecycleHandler) TearDownEnvironment(currentNode node.Node) error {
	return nil
}

// Start starts monitoring agents and delegates to another function to start blockchain containers
func (d DockerLifecycleHandler) Start(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	monitoringPath := client.AddBasePath("monitoring")
	filebeatCombinedConfigPath := client.AddBasePath(path.Join("monitoring", filebeatConfigFile))

	// Start filebeat container
	filebeatContainer := docker.Container{
		Name:  filebeatContainerName,
		Image: filebeatContainerImage,
		Cmd:   []string{"-e", "-strict.perms=false"},
		// using the first containers network is a decent default, if we ever do mult-network deployments we may need to rethink this
		Mounts: []docker.Mount{
			{
				Type: "bind",
				From: filebeatCombinedConfigPath,
				To:   "/usr/share/filebeat/filebeat.yml",
			},
			{
				Type: "bind",
				From: "/var/lib/docker/containers",
				To:   "/var/lib/docker/containers",
			},
			{
				Type: "bind",
				From: monitoringPath,
				To:   "/monitoring",
			},
			{
				Type: "bind",
				From: "/var/run/docker.sock",
				To:   "/var/run/docker.sock",
			},
		},
		User: "root",
	}

	if err := client.ContainerRuns(ctx, filebeatContainer); err != nil {
		return err
	}

	// Next, start the node containers
	for _, container := range d.containers {
		if err := client.ContainerRuns(ctx, container); err != nil {
			return err
		}
	}

	return nil
}

// Status returns the status of the running blockchain client and monitoring containers
func (d DockerLifecycleHandler) Status(currentNode node.Node) (string, error) {
	client, err := docker.NewBasicManager(currentNode)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	exists, err := client.DoesNetworkExist(ctx, currentNode.StrParameters["docker-network"])
	if err != nil {
		return "", err
	}
	if !exists {
		return "incomplete", nil
	}

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

// Stop removes all containers
func (d DockerLifecycleHandler) Stop(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for _, container := range d.containers {
		if err = client.ContainerStopped(ctx, container); err != nil {
			return err
		}
	}

	filebeatContainer := docker.Container{
		Name: filebeatContainerName,
	}

	if err = client.ContainerStopped(ctx, filebeatContainer); err != nil {
		return err
	}

	return nil
}

// RemoveData removes any data (typically the blockchain itself) related to the node
func (d DockerLifecycleHandler) RemoveData(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode)
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

	dataDir := client.AddBasePath(currentNode.StrParameters["data-dir"])
	fmt.Printf("Removing directory %q\n", dataDir)

	return os.RemoveAll(dataDir)
}

// RemoveRuntime removes the docker network and containers
func (d DockerLifecycleHandler) RemoveRuntime(currentNode node.Node) error {
	client, err := docker.NewBasicManager(currentNode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()

	for _, container := range d.containers {
		if err = client.ContainerAbsent(ctx, container); err != nil {
			return err
		}
	}

	filebeatContainer := docker.Container{
		Name: filebeatContainerName,
	}

	if err = client.ContainerAbsent(ctx, filebeatContainer); err != nil {
		return err
	}

	return nil
}
