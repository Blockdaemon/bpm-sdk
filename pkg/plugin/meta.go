package plugin

import (
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
