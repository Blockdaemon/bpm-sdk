package plugin

import (
	"gopkg.in/yaml.v2"
)

type Parameters struct {
	Network     []string
	Protocol    []string
	Subtype     []string
	NetworkType []string
}

func (p Parameters) String() string {
	d, err := yaml.Marshal(&p)
	if err != nil {
		panic(err) // Should never happen
	}

	return string(d)
}
