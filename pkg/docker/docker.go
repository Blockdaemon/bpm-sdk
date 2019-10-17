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
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type BasicManager struct {
	cli *client.Client
	prefix string
	basePath string
}

// NewBasicManager creates a BasicManager
//
// prefix is a string that gets added to every container-, network-, volume-name, etc. started by this client
// basePath is a path that gets added to every relative file paths. Example with basePath = /home/user/.bpm/nodes/xyz/config: test.yml becomes /home/user/.bpm/nodes/xyz/config/test.yml
func NewBasicManager(prefix, basePath string) (*BasicManager, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &BasicManager{
		cli: cli,
		prefix: prefix,
		basePath: basePath,
	}, nil
}

func (bm *BasicManager) prefixedName(name string) string {
	// make sure we don't accidentally double-prefix it
	if strings.HasPrefix(name, bm.prefix) {
		return name
	}

	return bm.prefix + name
}

func (bm *BasicManager) addBasePath(myPath string) string {
	if strings.HasPrefix(myPath, "/") {
		// absolute path, just return as is
		return myPath
	}

	return path.Join(bm.basePath, myPath)
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

// ContainerAbset stops and removes a container if it is running/exists
func (bm *BasicManager) ContainerAbsent(ctx context.Context, containerName string) error {
	running, err := bm.IsContainerRunning(ctx, containerName)
	if err != nil {
		return err
	}

	if running {
		fmt.Printf("Stopping container '%s'\n", containerName)

		if err := bm.cli.ContainerStop(ctx, bm.prefixedName(containerName), nil); err != nil {
			return err
		}
	} else {
		fmt.Printf("Container '%s' is not running, skipping stop\n", bm.prefixedName(containerName))
	}

	exists, err := bm.doesContainerExist(ctx, containerName)
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Removing container '%s'\n", bm.prefixedName(containerName))

		if err := bm.cli.ContainerRemove(ctx, bm.prefixedName(containerName), types.ContainerRemoveOptions{RemoveVolumes: true}); err != nil {
			return err
		}
	} else {
		fmt.Printf("Cannot find container '%s', skipping removel\n", containerName)
	}

	return nil
}

// NetworkAbsent removes a network if it exists
func (bm *BasicManager) NetworkAbsent(ctx context.Context, networkID string) error {
	exists, err := bm.doesNetworkExist(ctx, networkID)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("Cannot find network '%s', skipping removal\n", bm.prefixedName(networkID))
		return nil
	}

	fmt.Printf("Removing network '%s'\n", bm.prefixedName(networkID))
	return bm.cli.NetworkRemove(ctx, bm.prefixedName(networkID))
}

// VolumeAbsent removes a network if it exists
func (bm *BasicManager) VolumeAbsent(ctx context.Context, volumeID string) error {
	exists, err := bm.doesVolumeExist(ctx, volumeID)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("Cannot find volume '%s', skipping removal\n", bm.prefixedName(volumeID))
		return nil
	}

	fmt.Printf("Removing volume '%s'\n", bm.prefixedName(volumeID))
	return bm.cli.VolumeRemove(ctx, bm.prefixedName(volumeID), false)
}

// NetworkExists creates a network if it doesn't exist yet
func (bm *BasicManager) NetworkExists(ctx context.Context, networkID string) error {
	exists, err := bm.doesNetworkExist(ctx, networkID)
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Network '%s' already exists, skipping creation\n", bm.prefixedName(networkID))
		return nil
	}

	fmt.Printf("Creating network '%s'\n", bm.prefixedName(networkID))
	_, err = bm.cli.NetworkCreate(ctx, bm.prefixedName(networkID), types.NetworkCreate{CheckDuplicate: true})

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
	NetworkID   string
	EnvFilename string
	Mounts      []Mount
	Ports       []Port
	Cmd         []string
	CmdFile     string
	User        string
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

	if !exists {
		fmt.Printf("Creating container '%s'\n", bm.prefixedName(container.Name))

		if err := bm.createContainer(ctx, container); err != nil {
			return err
		}
	} else {
		fmt.Printf("Container '%s' already exists, skipping creation\n", bm.prefixedName(container.Name))
	}

	running, err := bm.IsContainerRunning(ctx, container.Name)
	if err != nil {
		return err
	}
	if !running {
		fmt.Printf("Starting container '%s'\n", bm.prefixedName(container.Name))

		if err := bm.cli.ContainerStart(ctx, bm.prefixedName(container.Name), types.ContainerStartOptions{}); err != nil {
			return err
		}
	} else {
		fmt.Printf("Container '%s' already runs, skipping start\n", bm.prefixedName(container.Name))
	}

	return nil
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

func (bm *BasicManager) doesNetworkExist(ctx context.Context, networkID string) (bool, error) {
	_, err := bm.cli.NetworkInspect(ctx, bm.prefixedName(networkID))
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
		envs, err = readLines(container.EnvFilename)
		if err != nil {
			return err
		}
	}

	// Ports
	portBindings := make(map[nat.Port][]nat.PortBinding)

	for _, portParameter := range container.Ports {
		containerPort, err := nat.NewPort(portParameter.Protocol, portParameter.ContainerPort)
		if err != nil {
			return err
		}

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

		from := mountParam.From
		if mountParam.Type == "bind" {
			from = bm.addBasePath(mountParam.From)

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
	endpointsConfig[container.NetworkID] = &network.EndpointSettings{
		NetworkID: bm.prefixedName(container.NetworkID),
	}
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}

	// Command
	cmd := []string{}
	if len(container.Cmd) > 0 {
		cmd = container.Cmd
	} else if len(container.CmdFile) > 0 {
		cmdFileContent, err := ioutil.ReadFile(bm.addBasePath(container.CmdFile))
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
		Image: container.Image,
		Env:   envs,
		Cmd:   cmd,
		User:  container.User,
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
