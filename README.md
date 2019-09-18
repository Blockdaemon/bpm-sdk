# About BPM-SDK

[![bpm-sdk](https://godoc.org/github.com/Blockdaemon/bpm-sdk?status.svg)](https://godoc.org/github.com/Blockdaemon/bpm-sdk)

BPM is the Blockchain Package Manager by Blockdaemon. It allows easy and uniform deployment, maintenance and upgrades of blockchain nodes.

BPM itself provides the framework, the actual deployment is performed by plugins. Plugins are just binaries that provide a common set of
command line parameters (see [Plugin Lifecycle](#plugin-lifecycle))


While it is possible to write plugins without an SDK, it is
recommended and significantly easier to use the provided go SDK.

## Plugin Lifecycle

Each plugin is a binary file that provides certain command line parameters. The different ways of invoking the plugin binary using command line parameters form the plugin lifecycle.

As a general rule each plugin command tries to be idempotent:

> An operation is idempotent if the result of performing it once is exactly the same as the result of performing it repeatedly without any intervening actions. [[1]](https://docs.ansible.com/ansible/latest/reference_appendices/glossary.html)

This is typically implemented by first checking if an action has already been applied and only take action if it hasn't. This has a few benefits:

- It allows manual intervention. Example: A user can supply their own node secret. If the file already exists, bpm will not re-create it
- Commands can be run multiple times without causing weird side effects
- Implementing new plugins is simpler because one can just re-run the command while implementing it

The SDK provides a lot of helper functions to achieve this without overhead.

### Creation of a new node

Whenever a user runs: `bpm run <plugin>`, bpm internally calls the plugin with the following parameters:

	plugin-binary create-secrets <node-id>

The `create-secrets` command is invoked first. It creates all secrets necessary for a node. This usually is a single node private key in `~/.blockdaemon/nodes/<node-id>/secrets/`.

	plugin-binary create-configurations <node-id>

`create-configurations` creates all the necessary configuration files in `~/.blockdaemon/nodes/<node-id>/configs`. Please not that some of these configuration files may contain secrets that were previously created using `create-secrets`.

	plugin-binary start <node-id>

Starts the node. The current version of bpm supports docker so this will start one or more docker containers.

	plugin-binary version

The last step is to get the version of the plugin. `bpm` will write this version to `~/.blockdaemon/nodes/<node-id>/version`. This version can be used when doing an upgrade to know which migrations to run.

### Removal of a node

Whenever a user runs: `bpm remove <plugin>`, bpm internally calls the plugin with the following parameters:

	plugin-binary remove <node-id>

This will stop all node processes and remove all active resources like docker containers. It will not remove any data or configuration. To achieve this, the user
needs to run `bpm remove <plugin> --purge`. The `--purge` parameter will be passed through to the plugin like this:

	plugin-binary remove <node-id> --purge

In this case the assumption is that the plugin also removes data and configuration files. It should not remove any secrets. The reason is that everything else can be re-created but secrets cannot.

### Upgrade

Whenever a user runs: `bpm upgrade <plugin>`, bpm internally calls the plugin with the following parameters:

	plugin-binary upgrade <node-id>

This will take all the necessary actions to upgrade a node. This could include stopping/restarting containers, running database migrations, re-creating configuration, etc.

After the upgrade `bpm` will write the current plugin version to `~/.blockdaemon/nodes/<node-id>/version`. This version can be used when doing another upgrade to know which migrations to run.

## Plugin SDK

### docker

Package docker provides a simple docker abstraction layer that makes it very easy to start and remove docker containers, networks and volumes.

For more information please refer to the [API documentation](./docs/docker.md).

### node

Package node provides an easy way to access node related information.

For more information please refer to the [API documentation](./docs/node.md).

### template

Package template implements functions to render Go templates to files using the node.Node struct as an imnput for the templates.

For more information please refer to the [API documentation](./docs/template.md).

### plugin

Package plugin provides an easy way to create the required CLI for a plugin. It abstracts away all the command line and file parsing so users just need to implement the actual logic. 

For more information please refer to [Implementing a plugin using the Go SDL](#implementing-a-plugin-using-the-go-sdk) and the [API documentation](./docs/plugin.md).

## Implementing a plugin using the Go SDK

The easiest way to get started is to copy the example below and start implementing the individual functions. Have a look at existing plugins for inspiration.

### Example with defaults

```go
package main

import (
	"fmt"

	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/Blockdaemon/bpm-sdk/pkg/plugin"
)

var version string

func start(currentNode node.Node) error {
	fmt.Println("Nothing to do here, skipping start")
	return nil
}

func main() {
	plugin.Initialize(plugin.Plugin{
		Name: "empty",
		Description: "A plugin that does nothing",
		Version: version,
		Start: start,
		CreateSecrets: plugin.DefaultCreateSecrets,
		CreateConfigs: plugin.DefaultCreateConfigs
		Remove: plugin.DefaultRemove,
		Upgrade: plugin.DefaultUpgrade,
	})
}
```

### Full Example

```go
package main

import (
	"fmt"

	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/Blockdaemon/bpm-sdk/pkg/plugin"
)

var version string

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
		Version: version,
		Start: start,
		CreateSecrets: createSecrets,
		CreateConfigs: createConfigs,
		Remove: remove,
		Upgrade: upgrade,
	})
}
```

### Tipps & Tricks

#### Automatically setting the version

Go has the capability to overwrite variables at compile time like this:

```bash
go build -ldflags "-X main.version=$VERSION" -o plugin-name cmd/main.go
```

This can be used in a continuous integration pipeline to automatically version the binaries with e.g. the current git tag.

## Nodestate

Every plugin should run nodestate next to the blockchain client. Nodestate sends monitoring information back to Blockdaemon so we can ensure the node is in perfect shape! Adapting nodestate to BPM is currently a work in progress. We'll update this section soon.

## Deployment

The plugin registry doesn't exist yet. In the meantime we use a digital ocean space to make plugins available. The uploaded files are the raw plugin binaries and filenames need to be in the following format:

`<plugin-name>-<version>-<operating-system>-<architecture>`

Example:

```
polkadot-0.5.1-linux-amd64
polkadot-0.5.1-darwin-amd64
```

## Continous Integration

A CI pipeline runs on https://gitlab.com/blockdaemon-public/bpm-sdk-ci-pipeline
