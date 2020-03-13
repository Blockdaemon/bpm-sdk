// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
package plugin

import (
	"github.com/Blockdaemon/bpm-sdk/pkg/fileutil"
	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/Blockdaemon/bpm-sdk/pkg/template"
)

// FileConfigurator creates configuration files from templates
type FileConfigurator struct {
	configFilesAndTemplates map[string]string
}

// Configure creates configuration files for the blockchain client
func (d FileConfigurator) Configure(currentNode node.Node) error {
	// Create config directory if it doesn't exist yet
	_, err := fileutil.MakeDirectory(currentNode.ConfigsDirectory())
	if err != nil {
		return err
	}

	return template.ConfigFilesRendered(d.configFilesAndTemplates, template.TemplateData{
		Node: currentNode,
	})
}

// RemoveConfig removes configuration files related to the node
func (d FileConfigurator) RemoveConfig(currentNode node.Node) error {
	// Remove all configuration files
	for file := range d.configFilesAndTemplates {
		if err := template.ConfigFileAbsent(file, currentNode); err != nil {
			return err
		}
	}

	return nil
}

// NewFileConfigurator creates an instance of FileConfigurator
func NewFileConfigurator(configFilesAndTemplates map[string]string) FileConfigurator {
	return FileConfigurator{
		configFilesAndTemplates: configFilesAndTemplates,
	}
}
