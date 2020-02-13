// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
package plugin

import (
	"fmt"

	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/Blockdaemon/bpm-sdk/pkg/template"
)

// FileConfigurator creates configuration files from templates
type FileConfigurator struct {
	configFilesAndTemplates map[string]string
	pluginParameters        []Parameter
}

// CreateSecrets does nothing except printing that it does nothing
func (d FileConfigurator) CreateSecrets(currentNode node.Node) error {
	if err := d.ValidateParameters(currentNode); err != nil {
		return err
	}
	fmt.Println("Nothing to do here, skipping create-secrets")
	return nil
}

// Configure creates configuration files for the blockchain client
func (d FileConfigurator) Configure(currentNode node.Node) error {
	if err := d.ValidateParameters(currentNode); err != nil {
		return err
	}
	return template.ConfigFilesRendered(d.configFilesAndTemplates, template.TemplateData{
		Node: currentNode,
	})
}

// Removes configuration files related to the node
func (d FileConfigurator) RemoveConfig(currentNode node.Node) error {
	// Remove all configuration files
	for file := range d.configFilesAndTemplates {
		if err := template.ConfigFileAbsent(file, currentNode); err != nil {
			return err
		}
	}

	return nil
}

// ValidateParameters checks if all specified parameters are provided
func (d FileConfigurator) ValidateParameters(currentNode node.Node) error {
	for _, parameter := range d.pluginParameters {
		ok := false

		if parameter.Type == ParameterTypeBool {
			_, ok = currentNode.BoolParameters[parameter.Name]
		}

		if parameter.Type == ParameterTypeString {
			_, ok = currentNode.StrParameters[parameter.Name]

		}

		if !ok {
			return fmt.Errorf(`%q missing`, parameter.Name)
		}
	}

	return nil
}

func NewFileConfigurator(configFilesAndTemplates map[string]string, pluginParameters []Parameter) FileConfigurator {
	return FileConfigurator{
		configFilesAndTemplates: configFilesAndTemplates,
		pluginParameters:        pluginParameters,
	}
}
