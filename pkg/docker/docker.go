// Package docker provides a simple docker abstraction layer.
//
// Please note that the methods are idempotent (i.e. they can be called multiple times without changing the result).
// This is important because it reduces the need for additional checks if the user runs a command multiple times. E.g.
// the code that uses this package doesn't need to check if the container already runs, ContainerRuns does that internally
// and just does nothing if the container is already running.
//
// Additionally it sometimes makes error handling simpler. If an particular method failed halfway, it can just be called
// again without causing any issues.
//
// The general pattern used internally in this package is:
//
// 		1. Check if the desired result (e.g. container running) already exists
// 		2. If yes, do nothing
// 		3. If no, invoke the action that produces the result (e.g. run container)
package docker

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.blockdaemon.com/bpm/sdk/pkg/node"
	sdktemplate "go.blockdaemon.com/bpm/sdk/pkg/template"
)

type BasicManager struct {
	cli         *client.Client
	currentNode node.Node
}

// NewBasicManager creates a BasicManager
func NewBasicManager(currentNode node.Node) (*BasicManager, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &BasicManager{
		cli:         cli,
		currentNode: currentNode,
	}, nil
}

func (bm *BasicManager) prefixedName(name string) string {
	// make sure we don't accidentally double-prefix it
	if strings.HasPrefix(name, bm.currentNode.NamePrefix()) {
		return name
	}

	return bm.currentNode.NamePrefix() + name
}

// AddBasePath adds the base path if the supplied path is relative
func (bm *BasicManager) AddBasePath(myPath string) string {
	if strings.HasPrefix(myPath, "/") {
		// absolute path, just return as is
		return myPath
	}

	return path.Join(bm.currentNode.NodeDirectory(), myPath)
}

// ListContainerNames lists all containers by name
func (bm *BasicManager) ListContainerNames(ctx context.Context) ([]string, error) {
	containers, err := bm.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return nil, err
	}

	names := []string{}

	for _, container := range containers {
		names = append(names, container.Names...) // The ... "unpacks" the Names array to merge it with names
	}

	// Docker names have a "/" in front of them, this package expects them not to have that so we'll remove it
	cleanNames := []string{}
	for _, name := range names {
		cleanNames = append(cleanNames, name[1:])
	}

	return cleanNames, nil
}

// ListVolumeIDs lists all volumes by name (which is also a unique id)
func (bm *BasicManager) ListVolumeIDs(ctx context.Context) ([]string, error) {
	volumesListOKBody, err := bm.cli.VolumeList(ctx, filters.Args{})
	if err != nil {
		return nil, err
	}

	names := []string{}

	for _, volume := range volumesListOKBody.Volumes {
		names = append(names, volume.Name)
	}

	return names, nil
}

// ContainerStopped stops a container if it is running
func (bm *BasicManager) ContainerStopped(ctx context.Context, container Container) error {
	prefixedName := bm.prefixedName(container.Name)

	running, err := bm.IsContainerRunning(ctx, container.Name)
	if err != nil {
		return err
	}

	if running {
		fmt.Printf("Stopping container '%s'\n", prefixedName)

		if err := bm.cli.ContainerStop(ctx, prefixedName, nil); err != nil {
			return err
		}
	} else {
		fmt.Printf("Container '%s' is not running, skipping stop\n", prefixedName)
	}

	return nil
}

// ContainerAbsent stops and removes a container if it is running/exists
func (bm *BasicManager) ContainerAbsent(ctx context.Context, container Container) error {
	prefixedName := bm.prefixedName(container.Name)

	if err := bm.ContainerStopped(ctx, container); err != nil {
		return err
	}

	exists, err := bm.doesContainerExist(ctx, container.Name)
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Removing container '%s'\n", prefixedName)

		if err := bm.cli.ContainerRemove(ctx, prefixedName, types.ContainerRemoveOptions{RemoveVolumes: true}); err != nil {
			return err
		}
	} else {
		fmt.Printf("Cannot find container '%s', skipping removel\n", prefixedName)
	}

	return nil
}

// NetworkAbsent removes a network if it exists
func (bm *BasicManager) NetworkAbsent(ctx context.Context, networkID string) error {
	exists, err := bm.DoesNetworkExist(ctx, networkID)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("Cannot find network '%s', skipping removal\n", networkID)
		return nil
	}

	fmt.Printf("Removing network '%s'\n", networkID)
	return bm.cli.NetworkRemove(ctx, networkID)
}

// VolumeAbsent removes a network if it exists
func (bm *BasicManager) VolumeAbsent(ctx context.Context, volumeID string) error {
	exists, err := bm.doesVolumeExist(ctx, volumeID)
	if err != nil {
		return err
	}

	prefixedName := bm.prefixedName(volumeID)

	if !exists {
		fmt.Printf("Cannot find volume '%s', skipping removal\n", prefixedName)
		return nil
	}

	fmt.Printf("Removing volume '%s'\n", prefixedName)
	return bm.cli.VolumeRemove(ctx, prefixedName, false)
}

// NetworkExists creates a network if it doesn't exist yet
func (bm *BasicManager) NetworkExists(ctx context.Context, networkID string) error {
	exists, err := bm.DoesNetworkExist(ctx, networkID)
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Network '%s' already exists, skipping creation\n", networkID)
		return nil
	}

	fmt.Printf("Creating network '%s'\n", networkID)
	_, err = bm.cli.NetworkCreate(ctx, networkID, types.NetworkCreate{CheckDuplicate: true})

	return err
}

// Mount defines a docker volume mount
type Mount struct {
	Type string
	From string
	To   string
}

// Port defines a forwarded docker port
type Port struct {
	HostIP        string
	HostPort      string
	ContainerPort string
	Protocol      string
}

// Container defines all parameters used to create a container
type Container struct {
	Name        string
	Image       string
	EnvFilename string
	Mounts      []Mount
	Ports       []Port
	Cmd         []string
	CmdFile     string
	User        string
	CollectLogs bool
}

// ContainerRuns creates and starts a container if it doesn't exist/run yet
func (bm *BasicManager) ContainerRuns(ctx context.Context, container Container) error {
	if err := bm.pullImage(ctx, container.Image); err != nil {
		return err
	}

	exists, err := bm.doesContainerExist(ctx, container.Name)
	if err != nil {
		return err
	}

	prefixedName := bm.prefixedName(container.Name)

	if !exists {
		fmt.Printf("Creating container '%s'\n", prefixedName)

		if err := bm.createContainer(ctx, container); err != nil {
			return err
		}
	} else {
		fmt.Printf("Container '%s' already exists, skipping creation\n", prefixedName)
	}

	running, err := bm.IsContainerRunning(ctx, container.Name)
	if err != nil {
		return err
	}
	if !running {
		fmt.Printf("Starting container '%s'\n", prefixedName)

		if err := bm.cli.ContainerStart(ctx, prefixedName, types.ContainerStartOptions{}); err != nil {
			return err
		}
	} else {
		fmt.Printf("Container '%s' already runs, skipping start\n", prefixedName)
	}

	return nil
}

// RunTransientContainer runs a container once and removes it after it is finished.
func (bm *BasicManager) RunTransientContainer(ctx context.Context, container Container) (string, error) {
	// See: https://docs.docker.com/develop/sdk/examples/

	if err := bm.pullImage(ctx, container.Image); err != nil {
		return "", err
	}

	exists, err := bm.doesContainerExist(ctx, container.Name)
	if err != nil {
		return "", err
	}

	prefixedName := bm.prefixedName(container.Name)

	if !exists {
		fmt.Printf("Creating container '%s'\n", prefixedName)

		if err := bm.createContainer(ctx, container); err != nil {
			return "", err
		}
	} else {
		fmt.Printf("Container '%s' already exists, skipping creation\n", prefixedName)
	}

	running, err := bm.IsContainerRunning(ctx, container.Name)
	if err != nil {
		return "", err
	}
	if !running {
		fmt.Printf("Starting container '%s'\n", prefixedName)

		if err := bm.cli.ContainerStart(ctx, prefixedName, types.ContainerStartOptions{}); err != nil {
			return "", err
		}
	} else {
		fmt.Printf("Container '%s' already runs, skipping start\n", prefixedName)
	}

	defer func() {
		// Removing the container after it's done
		if err := bm.ContainerAbsent(ctx, container); err != nil {
			panic(err)
		}
	}()

	status, err := bm.cli.ContainerWait(ctx, prefixedName)
	if err != nil {
		return "", err
	}

	// Get stdout and stderr
	outReader, err := bm.cli.ContainerLogs(ctx, prefixedName, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return "", err
	}
	defer outReader.Close()
	output, err := ioutil.ReadAll(outReader)
	outputStr := string(output)
	if err != nil {
		return outputStr, err
	}

	if status != 0 {
		return outputStr, fmt.Errorf("Container '%s' failed with status code: %d", prefixedName, status)
	}

	return outputStr, nil
}

func (bm *BasicManager) doesContainerExist(ctx context.Context, containerName string) (bool, error) {
	_, err := bm.cli.ContainerInspect(ctx, bm.prefixedName(containerName))
	if err != nil {
		if client.IsErrContainerNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// DoesNetworkExist returns true if the docker network exists
func (bm *BasicManager) DoesNetworkExist(ctx context.Context, networkID string) (bool, error) {
	_, err := bm.cli.NetworkInspect(ctx, networkID)
	if err != nil {
		if client.IsErrNetworkNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (bm *BasicManager) doesVolumeExist(ctx context.Context, volumeID string) (bool, error) {
	_, err := bm.cli.VolumeInspect(ctx, bm.prefixedName(volumeID))
	if err != nil {
		if client.IsErrVolumeNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (bm *BasicManager) IsContainerRunning(ctx context.Context, containerName string) (bool, error) {
	inspect, err := bm.cli.ContainerInspect(ctx, bm.prefixedName(containerName))
	if err != nil {
		if client.IsErrContainerNotFound(err) {
			return false, nil // a non existing container is not running!
		}

		return false, err
	}

	return inspect.State.Running, nil
}

func (bm *BasicManager) pullImage(ctx context.Context, imageName string) error {
	out, err := bm.cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := ioutil.ReadAll(out); err != nil {
		return err
	}

	return nil
}

func (bm *BasicManager) createContainer(ctx context.Context, container Container) error {
	// Environment variables
	var envs []string
	var err error

	if container.EnvFilename != "" {
		envs, err = readLines(bm.AddBasePath(container.EnvFilename))
		if err != nil {
			return err
		}
	}

	// Ports
	portBindings := make(map[nat.Port][]nat.PortBinding)
	exposedPorts := make(nat.PortSet)

	for _, portParameter := range container.Ports {
		containerPort, err := nat.NewPort(portParameter.Protocol, portParameter.ContainerPort)
		if err != nil {
			return err
		}

		exposedPorts[containerPort] = struct{}{}

		portBindings[containerPort] = []nat.PortBinding{
			{
				HostIP:   portParameter.HostIP,
				HostPort: portParameter.HostPort,
			},
		}
	}

	// Mountpoints
	var mounts []mount.Mount
	for _, mountParam := range container.Mounts {

		// Render the from parameter as template. This allows us to parameterize where things are stored
		// E.g.: "{{ .Node.StrParametrs.data-dir }}/my-special-data"
		tmpl, err := template.New("").Parse(mountParam.From)
		if err != nil {
			return err
		}
		output := bytes.NewBufferString("")
		if err := tmpl.Execute(output, sdktemplate.TemplateData{Node: bm.currentNode}); err != nil {
			return err
		}
		from := output.String()

		// If it is a volume we add a prefix to be able to identify it again
		// If it is a bind without '/' we assume it's relative to the node directory
		if mountParam.Type == "bind" {
			from = bm.AddBasePath(from)
		} else { // volume
			from = bm.prefixedName(from)
		}

		mounts = append(mounts, mount.Mount{
			Type:   mount.Type(mountParam.Type),
			Source: from,
			Target: mountParam.To,
		})
	}

	// Host config
	hostCfg := &dockercontainer.HostConfig{
		Mounts:       mounts,
		PortBindings: portBindings,
		RestartPolicy: dockercontainer.RestartPolicy{
			Name: "unless-stopped",
		},
		LogConfig: dockercontainer.LogConfig{
			Type: "json-file",
			Config: map[string]string{
				"max-size": "10m",
				"max-file": "3",
			},
		},
	}

	// Network config
	endpointsConfig := make(map[string]*network.EndpointSettings)
	endpointsConfig[bm.currentNode.StrParameters["docker-network"]] = &network.EndpointSettings{
		NetworkID: bm.currentNode.StrParameters["docker-network"],
	}
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}

	// Command
	cmd := []string{}
	if len(container.Cmd) > 0 {
		cmd = container.Cmd
	} else if len(container.CmdFile) > 0 {
		cmdFileContent, err := ioutil.ReadFile(bm.AddBasePath(container.CmdFile))
		if err != nil {
			return err
		}

		for _, parameter := range strings.Split(string(cmdFileContent), "\n") {
			if len(parameter) > 0 {
				cmd = append(cmd, strings.TrimSpace(parameter))
			}
		}
	}

	// Container config
	containerCfg := &dockercontainer.Config{
		Image:        container.Image,
		Env:          envs,
		Cmd:          cmd,
		User:         container.User,
		ExposedPorts: exposedPorts,
	}

	// Create a container with configs
	_, err = bm.cli.ContainerCreate(ctx, containerCfg, hostCfg, networkConfig, bm.prefixedName(container.Name))
	if err != nil {
		return err
	}

	return nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
