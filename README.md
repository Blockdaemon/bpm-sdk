# About BPM

# Plugin Lifecycle / Contract

Return codes

# Implementing a plugin using the Go SDK

## Example

	package main

	import (
		"gitlab.com/Blockdaemon/blockchain/bpm-lib/pkg/plugin"
		"gitlab.com/Blockdaemon/blockchain/bpm-lib/pkg/node"

		"fmt"
	)

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
			Version: "0.5.0",
			CreateSecrets: createSecrets,
			CreateConfigs: createConfigs,
			Start: start,
			Remove: remove,
			Upgrade: upgrade,
		})
	}


## How to build

TODO: Add specifics about version

# Nodestate