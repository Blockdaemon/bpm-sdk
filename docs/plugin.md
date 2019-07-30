# plugin
--
    import "github.com/Blockdaemon/bpm-sdk/pkg/plugin"

Package plugin provides an easy way to create the required CLI for a plugin. It
abstracts away all the command line and file parsing so users just need to
implement the actual logic.

Please see the main BPM-SDK documentation for more details on how to implement a
new plugin.

Usage Example:

    package main

    import (
    	"github.com/Blockdaemon/bpm-sdk/pkg/plugin"
    	"github.com/Blockdaemon/bpm-sdk/pkg/node"

    	"fmt"
    )

    var pluginVersion string

    func createSecrets(currentNode node.Node) error {
    	fmt.Println("Nothing to do here, skipping create-secrets")
    	return nil
    }

    func createConfigs(currentNode node.Node) error {
    	fmt.Println("Nothing to do here, skipping create-configurations")
    	return nil
    }

    func start(currentNode node.Node) error {
    	fmt.Println("Nothing to do here, skipping start")
    	return nil
    }

    func remove(currentNode node.Node, purge bool) error {
    	fmt.Println("Nothing to do here, skipping remove")
    	return nil
    }

    func upgrade(currentNode node.Node) error {
    	fmt.Println("Nothing to do here, skipping upgrade")
    	return nil
    }

    func main() {
    	plugin.Initialize(plugin.Plugin{
    		Name: "empty",
    		Description: "A plugin that does nothing",
    		Version: pluginVersion,
    		CreateSecrets: createSecrets,
    		CreateConfigs: createConfigs,
    		Start: start,
    		Remove: remove,
    		Upgrade: upgrade,
    	})
    }

## Usage

#### func  Initialize

```go
func Initialize(plugin Plugin)
```
Initialize creates the CLI for a plugin

#### type Plugin

```go
type Plugin struct {
	// The name of the plugin
	Name string
	// A short one-line description of the plugin
	Description string
	// The semantic version of the plugin. Please increment with every change to the plugin
	Version string
	// Function that creates the secrets for a node
	CreateSecrets func(currentNode node.Node) error
	// Function that creates the configuration for the blockchain client
	CreateConfigs func(currentNode node.Node) error
	// Function to start the node. This usually involves creating a Docker network and starting containers
	Start func(currentNode node.Node) error
	// Function to remove a running node. This usually involves removing Docker resources and deleting generated configuration files
	Remove func(currentNode node.Node, purge bool) error
	// Function to upgrade a node with a new plugin version
	Upgrade func(currentNode node.Node) error
}
```

Plugin describes and provides the functionality for a plugin
