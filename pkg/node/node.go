// Package node provides an easy way to access node related information.
//
// Utility functions to generate names and directory paths encapsulate the package conventions.
// It is highly recommended to use this package when implementing a new package to achieve consistency
// across packages.
package node

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/Blockdaemon/bpm-sdk/internal/util"
)

// Node represents a blockchain node, it's configuration and related information
type Node struct {
	baseDir string

	// The global ID of this node
	ID string `json:"id"`

	// Which blockchain network to connect to (Example: mainnet, ropsten, ...)
	Environment string `json:"environment"`

	// Describes the type of this blockchain network (Examples: public, private)
	NetworkType string `json:"networkType"`

	// Describes the protocol of this node (Examples: bitcoin, ethereum, polkadot, ...)
	Protocol string `json:"protocol"`

	// Describes the specific type of this node (Examples: validator, watcher, ...)
	Subtype string `json:"subtype"`
	// Describes the protocol of this node (Examples: bitcoin, ethereum, polkadot, ...)

	// Describes the collection configuration
	Collection Collection `json:"collection"`

	// Specific configuration settings for this node
	Config map[string]interface{} `json:"config"`

	// Secrets (Example: Private keys)
	Secrets map[string]interface{} `json:"-"` // No json here, never serialize secrets!

	// The package version used to install this node (if installed yet)
	// This is useful to know in order to run migrations on upgrades.
	Version string `json:"version"`
}

// Collection is config for log and node data collection
type Collection struct {
	CA   string `json:"ca"`
	Cert string `json:"cert"`
	Host string `json:"host"`
	Key  string `json:"key"`
}

// DockerPrefix returns the prefix used as a convention when naming containers, volumes and networks
func (c Node) DockerPrefix() string {
	return fmt.Sprintf("bd-%s", c.ID)
}

// DockerNetworkName returns the recommended name for a docker network in which this node runs
func (c Node) DockerNetworkName() string {
	return c.DockerPrefix()
}

// ContainerName takes a simple name for a docker container and returns it formatted according to package conventions
func (c Node) ContainerName(containerName string) string {
	return c.DockerPrefix() + "-" + containerName
}

// VolumeName converts a name for a docker volume and returns it formatted according to package conventions
func (c Node) VolumeName(volumeName string) string {
	return c.DockerPrefix() + "-" + volumeName
}

// NodeDirectory returns the base directory under which all configuration, secrets and meta-data for this node is stored
func (c Node) NodeDirectory() string {
	return path.Join(c.baseDir, c.ID)
}

// NodeFile returns the filepath in which the base configuration as well as meta-data from the PBG is stored
func (c Node) NodeFile() string {
	return path.Join(c.NodeDirectory(), "node.json")
}

// ConfigsDirectorys returns the directory under which all configuration for the blockchain client is stored
func (c Node) ConfigsDirectory() string {
	return path.Join(c.NodeDirectory(), "configs")
}

// ConfigsDirectorys returns the directory under which all secrets for the blockchain client is stored
func (c Node) SecretsDirectory() string {
	return path.Join(c.NodeDirectory(), "secrets")
}

// Load all the data for a particular node and creates all required directories
func Load(baseDir, id string) (Node, error) {
	node := Node{
		baseDir: baseDir,
		ID:      id,
	}

	// Create node directories if they don't exist yet
	_, err := util.MakeDirectory(node.SecretsDirectory())
	if err != nil {
		return node, err
	}

	_, err = util.MakeDirectory(node.ConfigsDirectory())
	if err != nil {
		return node, err
	}

	// Load node data
	nodeData, err := ioutil.ReadFile(node.NodeFile())
	if err != nil {
		return node, err
	}

	if err = json.Unmarshal(nodeData, &node); err != nil {
		return node, err
	}

	// Load secrets
	node.Secrets = make(map[string]interface{})

	files, err := ioutil.ReadDir(node.SecretsDirectory())
	if err != nil {
		return node, err
	}

	for _, f := range files {
		if !f.IsDir() {
			secret, err := ioutil.ReadFile(path.Join(node.SecretsDirectory(), f.Name()))
			if err != nil {
				return node, err
			}

			node.Secrets[f.Name()] = string(secret)
		}
	}

	return node, nil
}
