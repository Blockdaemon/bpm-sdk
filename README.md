# About BPM-SDK

BPM is the Blockchain Package Manager by Blockdaemon. It allows easy and uniform deployment, maintenance and upgrades of blockchain nodes.

BPM itself provides the framework, the actual deployment is performed by plugins. Plugins are just binaries that provide a common set of
command line parameters (see [Plugin Lifecycle](#plugin-lifecycle))


While it is possible to write plugins without an SDK, it is
recommended and significantly easier to use the provided go SDK.

# Plugin Lifecycle

Each plugin is a binary file that provides certain command line parameters. The different ways of invoking the plugin binary using command line parameters form the plugin lifecycle.

As a general rule each plugin command tries to be idempotent:

> An operation is idempotent if the result of performing it once is exactly the same as the result of performing it repeatedly without any intervening actions. [[1]](https://docs.ansible.com/ansible/latest/reference_appendices/glossary.html)

This is typically implemented by first checking if an action has already been applied and only take action if it hasn't. This has a few benefits:

- It allows manual intervention. Example: A user can supply their own node secret. If the file already exists, bpm will not re-create it
- Commands can be run multiple times without causing weird side effects
- Implementing new plugins is simpler because one can just re-run the command while implementing it

The SDK provides a lot of helper functions to achieve this without overhead.

## Creation of a new node

Whenever a user runs: `bpm run <plugin>`, bpm internally calls the plugin with the following parameters:

	plugin-binary create-secrets <node-id>

The `create-secrets` command is invoked first. It creates all secrets necessary for a node. This usually is a single node private key in `~/.blockdaemon/nodes/<node-id>/secrets/`.

	plugin-binary create-configurations <node-id>

`create-configurations` creates all the necessary configuration files in `~/.blockdaemon/nodes/<node-id>/configs`. Please not that some of these configuration files may contain secrets that were previously created using `create-secrets`.

	plugin-binary start <node-id>

Starts the node. The current version of bpm supports docker so this will start one or more docker containers.

	plugin-binary version

The last step is to get the version of the plugin. `bpm` will write this version to `~/.blockdaemon/nodes/<node-id>/version`. This version can be used when doing an upgrade to know which migrations to run.

## Removal of a node

Whenever a user runs: `bpm remove <plugin>`, bpm internally calls the plugin with the following parameters:

	plugin-binary remove <node-id>

This will stop all node processes and remove all active resources like docker containers. It will not remove any data or configuration. To achieve this, the user
needs to run `bpm remove <plugin> --purge`. The `--purge` parameter will be passed through to the plugin like this:

	plugin-binary remove <node-id> --purge

In this case the assumption is that the plugin also removes data and configuration files. It should not remove any secrets. The reason is that everything else can be re-created but secrets cannot.

## Upgrade

Whenever a user runs: `bpm upgrade <plugin>`, bpm internally calls the plugin with the following parameters:

	plugin-binary upgrade <node-id>

This will take all the necessary actions to upgrade a node. This could include stopping/restarting containers, running database migrations, re-creating configuration, etc.

After the upgrade `bpm` will write the current plugin version to `~/.blockdaemon/nodes/<node-id>/version`. This version can be used when doing another upgrade to know which migrations to run.

# Plugin SDK

## docker

Package docker provides a simple docker abstraction layer that makes it very easy to start and remove docker containers, networks and volumes.

For more information please refer to the [API documentation](./docs/docker.md).

## node

Package node provides an easy way to access node related information.

For more information please refer to the [API documentation](./docs/node.md).

## template

Package template implements functions to render Go templates to files using the node.Node struct as an imnput for the templates.

For more information please refer to the [API documentation](./docs/template.md).

## plugin

Package plugin provides an easy way to create the required CLI for a plugin. It abstracts away all the command line and file parsing so users just need to implement the actual logic. 

For more information please refer to [Implementing a plugin using the Go SDL](#implementing-a-plugin-using-the-go-sdk) and the [API documentation](./docs/plugin.md).

# Implementing a plugin using the Go SDK

The easiest way to get started is to copy the example below and start implementing the individual functions. Have a look at existing plugins for inspiration.

## Example

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


## Tipps & Tricks

### Automatically setting the version

Go has the capability to overwrite variables at compile time like this:

	go build -ldflags "-X main.pluginVersion=$VERSION" -o binary cmd/main.go

This can be used in a continuous integration pipeline to automatically version the binaries with e.g. the current git tag.

# Nodestate

Every plugin should run nodestate next to the blockchain client. Nodestate sends monitoring information back to Blockdaemon so we can ensure the node is in perfect shape! Adapting nodestate to BPM is currently a work in progress. We'll update this section soon.

# Deployment

The plugin registry doesn't exist yet. In the meantime we use a digital ocean space to make plugins available. The uploaded files are the raw plugin binaries and filenames need to be in the following format:

	<plugin-name>-<version>-<operating-system>-<architecture>

Example:

	polkadot-0.5.1-linux-amd64
	polkadot-0.5.1-darwin-amd64

In addition to the plugin binaries a single `version-info.json` needs to exist in the space. This file contains information about the *latest* versions. It looks like this:
	
	{
	    "runner-version": "0.3.0",
	    "plugins": [
	        {
	            "name": "stellar",
	            "version": "0.4.2"
	        },
	        {
	            "name": "polkadot",
	            "version": "0.5.2"
	        }
	    ]
	}


When uploading a new plugin, the name and version need to be added to this file.

Important: Make sure the files are public!

# BPM-SDK development

The following section contains information for the development of the BPM-SDK itself.

## Generate SDK documentation

Generating the SDK documentation is currently a bit tricky for different reasons:

* `godoc` doesn't support modules yet (https://github.com/golang/go/issues/26827)
* `godoc -html` does not produce a nice output (https://github.com/golang/go/issues/2381)
* `godoc` doesn't like symbolic links in the `GOROOT`. This breaks it when using homebrew on OSX.

The recommended workarounds to use `wget` result in incorrect links if only the `bpm-sdk` parts of the documentation are exported.

Until the above issues have been resolved, we are using (godocdown)[https://github.com/robertkrimen/godocdown] to generated markdown instead of html. Unfortunately `godocdown` doesn't work with modules yet so we still have to "fake" the go path.

First, install `godocdown`: `go get github.com/robertkrimen/godocdown/godocdown`

Then follow these steps:

1. Change into the `bpm-sdk` directory: `cd bpm-sdk`
2. Create a folder structure that resembles the traditional go directory: `mkdir -p ~/go-mod/src/github.com/Blockdaemon`
3. Symlink `bpm-sdk` into the new directory: `ln -s $(pwd) $HOME/go-mod/src/github.com/Blockdaemon/bpm-sdk`
4. Set `GOPATH` to the new directory: `export GOPATH=$HOME/go-mod/`
5. Export all packages:
   ```
   godocdown github.com/Blockdaemon/bpm-sdk/pkg/docker > docs/docker.md
   godocdown github.com/Blockdaemon/bpm-sdk/pkg/node > docs/node.md
   godocdown github.com/Blockdaemon/bpm-sdk/pkg/template > docs/template.md
   godocdown github.com/Blockdaemon/bpm-sdk/pkg/plugin > docs/plugin.md
   ```






godocdown github.com/Blockdaemon/bpm-sdk/pkg/docker > docs/docker.md


