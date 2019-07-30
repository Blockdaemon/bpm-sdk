# docker
--
    import "github.com/Blockdaemon/bpm-sdk/pkg/docker"

Package docker provides a simple docker abstraction layer.

Please note that the methods are idempotent (i.e. they can be called multiple
times without changing the result). This is important because it reduces the
need for additional checks if the user runs a command multiple times. E.g. the
code that uses this package doesn't need to check if the container already runs,
ContainerRuns does that internally and just does nothing if the container is
already running.

Additionally it sometimes makes error handling simpler. If an particular method
failed halfway, it can just be called again without causing any issues.

The general pattern used internally in this package is:

    1. Check if the desired result (e.g. container running) already exists
    2. If yes, do nothing
    3. If no, invoke the action that produces the result (e.g. run container)

## Usage

#### type BasicManager

```go
type BasicManager struct {
}
```


#### func  NewBasicManager

```go
func NewBasicManager() (*BasicManager, error)
```
NewBasicManager creates a BasicManager

#### func (*BasicManager) ContainerAbsent

```go
func (bm *BasicManager) ContainerAbsent(ctx context.Context, containerName string) error
```
ContainerAbset stops and removes a container if it is running/exists

#### func (*BasicManager) ContainerRuns

```go
func (bm *BasicManager) ContainerRuns(ctx context.Context, container Container) error
```
ContainerRuns creates and starts a container if it doesn't exist/run yet

#### func (*BasicManager) NetworkAbsent

```go
func (bm *BasicManager) NetworkAbsent(ctx context.Context, networkID string) error
```
NetworkAbsent removes a network if it exists

#### func (*BasicManager) NetworkExists

```go
func (bm *BasicManager) NetworkExists(ctx context.Context, networkID string) error
```
NetworkExists creates a network if it doesn't exist yet

#### func (*BasicManager) VolumeAbsent

```go
func (bm *BasicManager) VolumeAbsent(ctx context.Context, volumeID string) error
```
VolumeAbsent removes a network if it exists

#### type Container

```go
type Container struct {
	Name        string
	Image       string
	NetworkID   string
	EnvFilename string
	Mounts      []Mount
	Ports       []Port
	Cmd         []string
}
```

Container defines all parameters used to create a container

#### type Mount

```go
type Mount struct {
	Type string
	From string
	To   string
}
```

Mount defines a docker volume mount

#### type Port

```go
type Port struct {
	HostIP        string
	HostPort      string
	ContainerPort string
	Protocol      string
}
```

Port defines a forwarded docker port
