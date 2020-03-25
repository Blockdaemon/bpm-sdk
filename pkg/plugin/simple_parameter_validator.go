package plugin

import (
	"fmt"

	"github.com/Blockdaemon/bpm-sdk/pkg/node"
)

// SimpleParameterValidator is a simple validator
//
// It checks if all parameters exist and if mandatory parameters have a value
type SimpleParameterValidator struct {
	pluginParameters []Parameter
}

// ValidateParameters checks if mandatory parameters are passed in
func (m SimpleParameterValidator) ValidateParameters(currentNode node.Node) error {
	for _, parameter := range m.pluginParameters {
		ok := false

		if parameter.Type == ParameterTypeBool {
			_, ok = currentNode.BoolParameters[parameter.Name]

			if !ok {
				return fmt.Errorf(`the parameter %q is missing`, parameter.Name)
			}
		}

		if parameter.Type == ParameterTypeString {
			value, ok := currentNode.StrParameters[parameter.Name]

			if !ok {
				return fmt.Errorf(`the parameter %q is missing`, parameter.Name)
			}

			if value == "" {
				if parameter.Mandatory {
					return fmt.Errorf(`the mandatory parameter %q is empty`, parameter.Name)
				}

				if parameter.Default != "" {
					return fmt.Errorf(`the parameter %q is empty but it should have a default`, parameter.Name)
				}
			}
		}

	}

	return nil
}

// NewSimpleParameterValidator creates an instance of SimpleParameterValidator
func NewSimpleParameterValidator(pluginParameters []Parameter) SimpleParameterValidator {
	return SimpleParameterValidator{
		pluginParameters: pluginParameters,
	}
}
