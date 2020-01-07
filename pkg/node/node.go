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
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Blockdaemon/bpm-sdk/internal/util"
	homedir "github.com/mitchellh/go-homedir"
)

// Node represents a blockchain node, it's configuration and related information
type Node struct {
	nodeFile string

	// The global ID of this node
	ID string `json:"id"`

	// The plugin name
	PluginName string `json:"plugin"`

	// Dynamic (i.e. defined by the plugin) string parameters
	StrParameters map[string]string `json:"str_parameters"`

	// Dynamic bool parameters
	BoolParameters map[string]bool `json:"bool_parameters"`

	// Describes the collection configuration
	Collection Collection `json:"collection"`

	// Secrets (Example: Private keys)
	Secrets map[string]interface{} `json:"-"` // No json here, never serialize secrets!

	// Holding place for data that is generated at runtime. E.g. can be used to store data parsed from the parameters
	Data map[string]interface{} `json:"-"` // No json here, runtime data only

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

// NamePrefix returns the prefix used as a convention when naming containers, volumes, networks, etc.
func (c Node) NamePrefix() string {
	return fmt.Sprintf("bpm-%s-", c.ID)
}

// NodeDirectory returns the base directory under which all configuration, secrets and meta-data for this node is stored
func (c Node) NodeDirectory() string {
	dir := filepath.Dir(c.nodeFile)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		panic(err) // Should never happen
	}

	expandedBaseDir, err := homedir.Expand(absDir)
	if err != nil {
		panic(err) // Should never happen
	}

	return expandedBaseDir
}

// NodeFile returns the filepath in which the base configuration as well as meta-data from the PBG is stored
func (c Node) NodeFile() string {
	return c.nodeFile
}

// ConfigsDirectorys returns the directory under which all configuration for the blockchain client is stored
func (c Node) ConfigsDirectory() string {
	return path.Join(c.NodeDirectory(), "configs")
}

// ConfigsDirectorys returns the directory under which all secrets for the blockchain client is stored
func (c Node) SecretsDirectory() string {
	return path.Join(c.NodeDirectory(), "secrets")
}

// Save the node data
func (c Node) Save() error {
	// Create node directories if they don't exist yet
	_, err := util.MakeDirectory(c.SecretsDirectory())
	if err != nil {
		return err
	}

	_, err = util.MakeDirectory(c.ConfigsDirectory())
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		c.NodeFile(),
		data,
		os.ModePerm,
	)
}

func New(nodeFile string) Node {
	return Node{nodeFile: nodeFile}
}

// Load all the data for a particular node and creates all required directories
func Load(nodeFile string) (Node, error) {
	node := New(nodeFile)

	// Load node data
	nodeData, err := ioutil.ReadFile(nodeFile)
	if err != nil {
		return node, err
	}

	if err = json.Unmarshal(nodeData, &node); err != nil {
		return node, err
	}

	// TODO: Using directories here as a shortcut. Not every plugin will use directories.
	//       E.g. if a plugin runs on k8s it might create k8s secrets.
	//       We will neeed to refactor this at some point!

	// Create node directories if they don't exist yet
	_, err = util.MakeDirectory(node.SecretsDirectory())
	if err != nil {
		return node, err
	}
	_, err = util.MakeDirectory(node.ConfigsDirectory())
	if err != nil {
		return node, err
	}

	// Initialize temporary data store
	node.Data = make(map[string]interface{})

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

			// as a convenience we parse json here so that individual elements
			// can be referenced when rendering templates
			if strings.HasSuffix(f.Name(), ".json") {
				var data interface{}
				if err := json.Unmarshal(secret, &data); err != nil {
					return node, err
				}

				node.Secrets[f.Name()] = data
			} else {
				node.Secrets[f.Name()] = string(secret)
			}
		}
	}

	return node, nil
}
