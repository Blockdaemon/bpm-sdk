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
	"path/filepath"

	"github.com/Blockdaemon/bpm-sdk/pkg/fileutil"
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

// NodeDirectory returns the base directory under which all configuration and meta-data for this node is stored
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

// Save the node data
func (c Node) Save() error {
	// Create node directories if they don't exist yet
	_, err := fileutil.MakeDirectory(c.NodeDirectory())
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

	// Initialize temporary data store
	node.Data = make(map[string]interface{})

	return node, nil
}
