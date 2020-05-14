package plugin

import (
	"github.com/coreos/go-semver/semver"
	"github.com/thoas/go-funk"
	"gopkg.in/yaml.v2"
)

const (
	ParameterTypeBool   = "bool"
	ParameterTypeString = "string"

	SupportsTest     = "test"
	SupportsUpgrade  = "upgrade"
	SupportsIdentity = "identity"
)

type Parameter struct {
	Type        string
	Name        string
	Description string
	Mandatory   bool
	Default     string
}

type MetaInfo struct {
	Name            string
	Version         string
	Description     string
	ProtocolVersion string `yaml:"protocol_version"`
	Parameters      []Parameter
	Supported       []string
}

func (p MetaInfo) String() string {
	d, err := yaml.Marshal(&p)
	if err != nil {
		panic(err) // Should never happen
	}

	return string(d)
}

// Supports returns bool if a particular method is supported
func (p MetaInfo) Supports(supported string) bool {
	return funk.ContainsString(p.Supported, supported)
}

// ProtocolVersionGreaterEqualThan return true if the protocol version is greater or equal to the provided version
func (p MetaInfo) ProtocolVersionGreaterEqualThan(version string) bool {
	v1 := semver.New(p.ProtocolVersion)
	v2 := semver.New(version)

	return v2.LessThan(*v1) || v2.Equal(*v1)
}
