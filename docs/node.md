# node
--
    import "gitlab.com/Blockdaemon/bpm-sdk/pkg/node"

Package node provides an easy way to access node related information.

Utility functions to generate names and directory paths encapsulate the plugin
conventions. It is highly recommended to use this package when implementing a
new plugin to achieve consistency across plugins.

## Usage

#### type Node

```go
type Node struct {
	// The global ID of this node
	NodeGID string `json:"node_gid"`

	// The global ID of the blockchain this node belongs to.
	// This value becomes very relevant when running a private/permissioned Blockchain network
	BlockchainGID string `json:"blockchain_gid"`

	// Which blockchain network to connect to (Example: mainnet, ropsten, ...)
	Environment string `json:"environment"`
	// Describes the type of this blockchain network (Examples: public, private)
	NetworkType string `json:"network_type"`
	// Describes the specific type of this node (Examples: validator, watcher, ...)
	NodeSubtype string `json:"node_subtype"`
	// Describes the protocol of this node (Examples: bitcoin, ethereum, polkadot, ...)
	ProtocolType string `json:"protocol_type"`

	// Specific configuration settings for this node
	Config map[string]interface{} `json:"config"`

	// Secrets (Example: Private keys)
	Secrets map[string]interface{} // No json here, never serialize secrets!

	// The plugin version used to install this node (if installed yet)
	// This is useful to know in order to run migrations on upgrades.
	CurrentVersion string
}
```

Node represents a blockchain node, it's configuration and related information

#### func  LoadNode

```go
func LoadNode(baseDir, nodeGID string) (Node, error)
```
LoadNode loads all the data for a particular node and creates all recommended
directories if they don't exist yet

#### func (Node) ConfigsDirectory

```go
func (c Node) ConfigsDirectory() string
```
ConfigsDirectorys returns the directory under which all configuration for the
blockchain client is stored

#### func (Node) ContainerName

```go
func (c Node) ContainerName(containerName string) string
```
ContainerName takes a simple name for a docker container and returns it
formatted according to plugin conventions

#### func (Node) CurrentVersionFile

```go
func (c Node) CurrentVersionFile() string
```
CurrentVersionFile returns the filepath in which the plugin version from the
last successfull install is stored

#### func (Node) DockerNetworkName

```go
func (c Node) DockerNetworkName() string
```
DockerNetworkName returns the recommended name for a docker network in which
this node runs

#### func (Node) NodeDirectory

```go
func (c Node) NodeDirectory() string
```
NodeDirectory returns the base directory under which all configuration, secrets
and meta-data for this node is stored

#### func (Node) NodeFile

```go
func (c Node) NodeFile() string
```
NodeFile returns the filepath in which the base configuration as well as
meta-data from the PBG is stored

#### func (Node) SecretsDirectory

```go
func (c Node) SecretsDirectory() string
```
ConfigsDirectorys returns the directory under which all secrets for the
blockchain client is stored

#### func (Node) VolumeName

```go
func (c Node) VolumeName(volumeName string) string
```
VolumeName converts a name for a docker volume and returns it formatted
according to plugin conventions

#### func (Node) WritePluginVersion

```go
func (c Node) WritePluginVersion(version string) error
```
WritePluginVersion writes the current plugin version into a version file. This
will be executed automatically by bpm after an new node is started or upgraded.
