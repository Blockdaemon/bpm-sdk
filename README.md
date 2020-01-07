# About BPM-SDK

[![bpm-sdk](https://godoc.org/github.com/Blockdaemon/bpm-sdk?status.svg)](https://godoc.org/github.com/Blockdaemon/bpm-sdk)

BPM is the Blockchain Package Manager by Blockdaemon. It allows easy and uniform deployment, maintenance and upgrades of blockchain nodes.

BPM itself provides the framework, the actual deployment is performed by plugins. Plugins are just binaries that provide a common set of
command line parameters (see [Plugin Lifecycle](#plugin-lifecycle))

While it is possible to write plugins without an SDK, it is recommended and significantly easier to use the provided go SDK.

## Plugin Lifecycle

Each plugin is a binary file that provides certain command line parameters. The different ways of invoking the plugin binary using command line parameters form the plugin lifecycle. Please note that an end-user never has to deal with the plugin command line parameters directly. Instead they will use the `bpm-cli` which internally calls the plugin via cli parameters.

As a general rule each plugin command tries to be idempotent:

> An operation is idempotent if the result of performing it once is exactly the same as the result of performing it repeatedly without any intervening actions. [[1]](https://docs.ansible.com/ansible/latest/reference_appendices/glossary.html)

This is typically implemented by first checking if an action has already been applied and only take action if it hasn't. This has a few benefits:

- It allows manual intervention. Example: A user can supply their own node secret. If the file already exists, bpm will not re-create it
- Commands can be run multiple times without causing weird side effects
- Implementing new plugins is simpler because one can just re-run the command while implementing it

The SDK provides a lot of helper functions to achieve this without overhead.

### Example: Starting a new node

Whenever a user runs: `bpm configure <plugin> [--network|--network-type|--subtype|--protocol|...]`, the `bpm-cli` does the following:

1. Checks if the plugin is up-to-date. If there is a newer version, it recommends to install it
2. Generates a random node id 
3. Writes plugin details and high level configuration to `~/.bpm/nodes/<node-id>/node.json`
4. Calls `plugin-binary create-secrets <node-id>`. This creates all secrets necessary for a node. This usually is a single node private key in `~/.bpm/nodes/<node-id>/secrets/`.
5. Calls `plugin-binary create-configurations <node-id>`. This creates all the necessary configuration files in `~/.bpm/nodes/<node-id>/configs`. Please not that some of these configuration files may contain secrets that were previously created using `create-secrets`. The configuration from `~/.bpm/nodes/<node-id>/node.json` is typically used to customize the protocol configuration files. For example if the user specifies `--subtype=validator`, this information shows up in the `node.json` and can be used to generate a protocol configuration specifically tailored for validators.

At this stage, the user has the ability to manually edit the protocol configuration in `~/.bpm/nodes/<node-id>/configs` if special configuration is needed.

Next the user can run `bpm start <node-id>` which internally calls `plugin-binary start <node-id>` to start the blockchain

There are similar commands for stopping, removing, etc. the node. Refer to the `bpm-cli` documentation for a full list.

### Lifecycle reference

**create-configurations**

Creates the configuration files. This is typically done by applying the contents of `node.json` to templates. This allows for different configuration depending on e.g. the `subtype`.

The plugin should write user friendly log messages to stdout and return with exit code 0 on success. On error it should write the error message to stderr and exit with a code other than 0.

**create-secrets**

Creates the secrets for a blockchain node and stores them on disk.

The plugin should write user friendly log messages to stdout and return with exit code 0 on success. On error it should write the error message to stderr and exit with a code other than 0.

**parameters**

Returns all valid values for the parameters used during the `bpm configure <plugin>` call. The values are returned in yaml format. The first value is always used as the default (even if there is only one value).

Example:

	$ polkadot parameters
	network:
	- alexander
	protocol:
	- polkadot
	subtype:
	- watcher
	- validator
	networktype:
	- public

The above example shows a polkadot plugin that supports the `watcher` and `validator` nodes on the `alexander` testnet. This will get expanded with more networks (e.g. `mainnet`) once they become available. This particular plugin supports only `polkadot` as a `protocol` and `public` as `networktype`.

The plugin should always exit out of this command with an exit code of 0.

**remove-config**

Removes all previously generated configurations.

The plugin should write user friendly log messages to stdout and return with exit code 0 on success. On error it should write the error message to stderr and exit with a code other than 0.

**remove-data**

Removes all node specific data (e.g. the blockchain itself) but not configuration. This can be useful to "reset" the node.

The plugin should write user friendly log messages to stdout and return with exit code 0 on success. On error it should write the error message to stderr and exit with a code other than 0.

**remove-node**

Removes the node itself. Depending on the target platform this can mean different things. If the plugin uses Docker it means removing the docker containers and networks

The plugin should write user friendly log messages to stdout and return with exit code 0 on success. On error it should write the error message to stderr and exit with a code other than 0.

**status**

Writes the status of the node to stdout. This can be one of:

- `running` if the node is currently running
- `stopped` if the node is currently stopped
- `incomplete` if parts of the node (e.g. only some containers) are running

The plugin should return with exit code 0 on success. On error it should write the error message to stderr and exit with a code other than 0.

**stop**

Stops the node so that it can later be started again. Behaviour depends on the target platform.

The plugin should write user friendly log messages to stdout and return with exit code 0 on success. On error it should write the error message to stderr and exit with a code other than 0.

**test**

Runs tests against the running node and reports the results on stdout. If at least one test failed it must return an exit code that is **not 0**.

**upgrade**

Upgrades an existing node to the current version of the plugin. This typically means stopping and restarting with a new version of the protocol. Sometimes other maintenance tasks need to be performed as well (.e.g database migrations).

The plugin should write user friendly log messages to stdout and return with exit code 0 on success. On error it should write the error message to stderr and exit with a code other than 0.

**version**

Writes the semantic version string of the plugin to stdout. Should always exit with exit code 0.

## Plugin SDK

### docker

Package docker provides a simple docker abstraction layer that makes it very easy to start and remove docker containers, networks and volumes.

For more information please refer to the [API documentation](https://godoc.org/github.com/Blockdaemon/bpm-sdk/pkg/docker).

### node

Package node provides an easy way to access node related information.

For more information please refer to the [API documentation](https://godoc.org/github.com/Blockdaemon/bpm-sdk/pkg/node).

### template

Package template implements functions to render Go templates to files using the node.Node struct as an imnput for the templates.

For more information please refer to the [API documentation](https://godoc.org/github.com/Blockdaemon/bpm-sdk/pkg/template).

### plugin

Package plugin provides an easy way to create the required CLI for a plugin. It abstracts away all the command line and file parsing so users just need to implement the actual logic. 

For more information please refer to [Implementing a plugin using the Go SDL](#implementing-a-docker-based-plugin-using-the-go-sdk) and the [API documentation](https://godoc.org/github.com/Blockdaemon/bpm-sdk/pkg/plugin).

## Implementing a Docker based plugin using the Go SDK

The easiest way to get started is to copy the example below and start changing the containers and settings.

### Example

```go
package main

import (
	"github.com/Blockdaemon/bpm-sdk/pkg/docker"
	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/Blockdaemon/bpm-sdk/pkg/plugin"
)

var version string

const (
	polkadotContainerImage = "docker.io/chevdor/polkadot:0.4.4"
	polkadotContainerName  = "polkadot"
	polkadotDataVolumeName = "polkadot-data"
	polkadotCmdFile        = "polkadot.dockercmd"

	networkName = "polkadot"

	polkadotCmdTpl = `polkadot
--base-path
/data
--rpc-external
--name
{{ .Node.ID }}
--chain
{{ .Node.Environment }}
{{ if eq .Node.Subtype "validator" }}
--validator
--key {% ADD NODE KEY HERE %}
{{ end }}
`
)

func main() {
	// Define the container that runs the blockchain
	containers := []docker.Container{
		{
			Name:      polkadotContainerName,
			Image:     polkadotContainerImage,
			CmdFile:   polkadotCmdFile,
			NetworkID: networkName,
			Mounts: []docker.Mount{
				{
					Type: "volume",
					From: polkadotDataVolumeName,
					To:   "/data",
				},
			},
			Ports: []docker.Port{
				{
					HostIP:        "0.0.0.0",
					HostPort:      "30333",
					ContainerPort: "30333",
					Protocol:      "tcp",
				},
				{
					HostIP:        "127.0.0.1",
					HostPort:      "9933",
					ContainerPort: "9933",
					Protocol:      "tcp",
				},
			},
			CollectLogs: true,
		},
	}

	// Define the allowed parameters
	parameters := plugin.Parameters{
		Network:     []string{"alexander"},
		Protocol:    []string{"polkadot"},
		Subtype:     []string{"watcher", "validator"},
		NetworkType: []string{"public"},
	}

	// Define templates
	templates := map[string]string{
		polkadotCmdFile: polkadotCmdTpl,
	}

	// Initialize the plugin
	dockerPlugin := plugin.NewDockerPlugin(
		"polkadot",
		"A polkadot plugin",
		version,
		containers,
		templates,
		parameters,
	)
	plugin.Initialize(dockerPlugin)
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
