package plugin

import (
	"go.blockdaemon.com/bpm/sdk/pkg/docker"
)

// DockerPlugin is an implementation of the Plugin interface. It provides based functionality for a docker based plugin
type DockerPlugin struct {
	ParameterValidator
	IdentityCreator
	Configurator
	LifecycleHandler
	Upgrader
	Tester

	// Plugin meta information
	meta MetaInfo
}

// Name returns the name of a plugin
func (d DockerPlugin) Name() string {
	return d.meta.Name
}

// Meta returns the MetaInfo of a plugin
func (d DockerPlugin) Meta() MetaInfo {
	// Determine optional functions available on the fly
	supported := []string{}

	if d.Tester != nil {
		supported = append(supported, SupportsTest)
	}

	if d.Upgrader != nil {
		supported = append(supported, SupportsUpgrade)
	}

	if d.IdentityCreator != nil {
		supported = append(supported, SupportsIdentity)
	}

	d.meta.Supported = supported

	return d.meta
}

// NewDockerPlugin creates a new instance of DockerPlugin
func NewDockerPlugin(name string, version string, description string, parameters []Parameter, templates map[string]string, containers []docker.Container) DockerPlugin {
	dockerParameters := []Parameter{
		{
			Name:        "docker-network",
			Type:        ParameterTypeString,
			Description: "If set, the node will be spun up in this docker network. The network will be created automatically if it doesn't exist",
			Mandatory:   false,
			Default:     "bpm",
		},
		{
			Name:        "data-dir",
			Type:        ParameterTypeString,
			Description: "The directory under which the nodes data will be saved. Values that do not start with '/' will be relative to the node directory",
			Mandatory:   false,
			Default:     "data",
		},
		{
			Name:        "monitoring-pack",
			Type:        ParameterTypeString,
			Description: "Enables sending monitoring data to an endpoint using settings from the monitoring pack (a *.tar.gz file)",
			Mandatory:   false,
			Default:     "",
		},
	}

	meta := MetaInfo{
		Name:            name,
		Version:         version,
		Description:     description,
		ProtocolVersion: "1.2.0",
		Parameters:      append(dockerParameters, parameters...),
		Supported:       []string{}, // We'll determine the supported functions on the fly in DockerPlugin.Meta()
	}

	return DockerPlugin{
		meta:               meta,
		ParameterValidator: NewSimpleParameterValidator(meta.Parameters),
		IdentityCreator:    nil,
		Configurator:       NewFileConfigurator(templates),
		LifecycleHandler:   NewDockerLifecycleHandler(containers),
		Upgrader:           NewDockerUpgrader(containers),
		Tester:             nil,
	}
}
