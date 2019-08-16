// Package node provides an easy way to access node related information.
//
// Utility functions to generate names and directory paths encapsulate the plugin conventions.
// It is highly recommended to use this package when implementing a new plugin to achieve consistency
// across plugins.
package node

import (
	"io/ioutil"
	"encoding/json"
	"path"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/Blockdaemon/bpm-sdk/internal/util"
)

// Node represents a blockchain node, it's configuration and related information
type Node struct {
	// The global ID of this node
	NodeGID string `json:"node_gid"`

	// The global ID of the blockchain this node belongs to. 
	// This value becomes very relevant when running a private/permissioned Blockchain network
	BlockchainGID string `json:"blockchain_gid"`

	// Which blockchain network to connect to (Example: mainnet, ropsten, ...)
	Environment  string `json:"environment"`
	// Describes the type of this blockchain network (Examples: public, private)
	NetworkType  string `json:"network_type"`
	// Describes the specific type of this node (Examples: validator, watcher, ...)
	NodeSubtype  string `json:"node_subtype"`
	// Describes the protocol of this node (Examples: bitcoin, ethereum, polkadot, ...)
	ProtocolType string `json:"protocol_type"`
	// Describes the logstash configuration
	Logstash LogstashConfig `json:"logstash"`

	// Specific configuration settings for this node
	Config map[string]interface{} `json:"config"`

	// Secrets (Example: Private keys)
	Secrets map[string]interface{} // No json here, never serialize secrets!

	// The plugin version used to install this node (if installed yet)
	// This is useful to know in order to run migrations on upgrades.
	CurrentVersion string

	baseDir string
}

// LogstashConfig represents the configuration for logstash
type LogstashConfig struct {
	Host string `json:"host"`
	Certificate string `json:"certificate"`
	CertificateAuthorities string `json:"certificate_authorities"`
	Key string `json:"key"`
}

// DockerPrefix returns the prefix used as a convention when naming containers, volumes and networks
func (c Node) DockerPrefix() string {
	return "bd-" + c.NodeGID
}

// DockerNetworkName returns the recommended name for a docker network in which this node runs
func (c Node) DockerNetworkName() string {
	return c.DockerPrefix()
}

// ContainerName takes a simple name for a docker container and returns it formatted according to plugin conventions
func (c Node) ContainerName(containerName string) string {
	return c.DockerPrefix() + "-" + containerName
}

// VolumeName converts a name for a docker volume and returns it formatted according to plugin conventions
func (c Node) VolumeName(volumeName string) string {
	return c.DockerPrefix() + "-" + volumeName
}

// NodeDirectory returns the base directory under which all configuration, secrets and meta-data for this node is stored
func (c Node) NodeDirectory() string {
	expandedBaseDir, err := homedir.Expand(c.baseDir)
	if err != nil {
		panic(err) // Should never happen because, at this stage, the directory should already be created
	}

	return path.Join(expandedBaseDir, "nodes", c.NodeGID)
}

// CurrentVersionFile returns the filepath in which the plugin version from the last successfull install is stored
func (c Node) CurrentVersionFile() string {
	return path.Join(c.NodeDirectory(), "version")
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

// WritePluginVersion writes the current plugin version into a version file. 
// This will be executed automatically by bpm after an new node is started or upgraded.
func (c Node) WritePluginVersion(version string) error {
	c.CurrentVersion = version

	return ioutil.WriteFile(c.CurrentVersionFile(), []byte(version), 0644)
}

// LoadNode loads all the data for a particular node and creates all recommended directories if they don't exist yet
func LoadNode(baseDir, nodeGID string) (Node, error) {
	var node Node

	node.NodeGID = nodeGID
	node.baseDir = baseDir

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

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
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

    // Load installed version (if there is any)
    exists, err := util.FileExists(node.CurrentVersionFile())
    if err != nil {
    	return node, err
    }

    if exists {
    	versionData, err := ioutil.ReadFile(node.CurrentVersionFile())
    	if err != nil {
    		return node, err
    	}

    	node.CurrentVersion = string(versionData)
    }

	return node, nil
}

