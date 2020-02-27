// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
//
// Please see the main BPM-SDK documentation for more details on how to implement a new plugin.
package plugin

import (
	"fmt"
	"os"

	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
)

// Configurator is the interface that wraps the Configure method
type Configurator interface {
	// Function that creates the configuration for the node
	Configure(currentNode node.Node) error

	// Removes configuration related to the node
	RemoveConfig(currentNode node.Node) error
}

// LifecycleHandler provides functions to manage a node
type LifecycleHandler interface {
	// Function to start a node
	Start(currentNode node.Node) error
	// Function to stop a running node
	Stop(currentNode node.Node) error
	// Function to return the status (running, incomplete, stopped) of a node
	Status(currentNode node.Node) (string, error)
	// Removes any data (typically the blockchain itself) related to the node
	RemoveData(currentNode node.Node) error
	// Removes everything other than data and configuration related to the node
	RemoveRuntime(currentNode node.Node) error
}

// Upgrader is the interface that wraps the Upgrade method
type Upgrader interface {
	// Function to upgrade a node with a new plugin version
	Upgrade(currentNode node.Node) error
}

// Tester is the interface that wraps the Test method
type Tester interface {
	// Function to test a node
	Test(currentNode node.Node) (bool, error)
}

// Plugin describes and provides the functionality for a plugin
type Plugin interface {
	// Returns the name of the plugin
	Name() string
	// Return plugin meta information such as: What's supported, possible parameters
	Meta() MetaInfo

	Configurator
	LifecycleHandler
	Upgrader
	Tester
}

type PluginImpl struct {
	Configurator
	LifecycleHandler
	Upgrader
	Tester

	// Plugin meta information
	meta MetaInfo

	name string
}

func (d PluginImpl) Name() string {
	return d.name
}

func (d PluginImpl) Meta() MetaInfo {
	return d.meta
}

func NewPlugin(name string, meta MetaInfo, configurator Configurator, lifecycleHandler LifecycleHandler, upgrader Upgrader, tester Tester) Plugin {
	return PluginImpl{
		name:             name,
		meta:             meta,
		Configurator:     configurator,
		LifecycleHandler: lifecycleHandler,
		Upgrader:         upgrader,
		Tester:           tester,
	}
}

// Initialize creates the CLI for a plugin
func Initialize(plugin Plugin) {
	// Initialize root command
	var rootCmd = &cobra.Command{
		Use:          plugin.Name(),
		Short:        plugin.Meta().Description,
		SilenceUsage: true,
	}

	// Create the commands
	var createConfigurationsCmd = &cobra.Command{
		Use:   "create-configurations <node-file>",
		Short: "Creates the configurations for a blockchain node and stores them on disk",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.Configure(currentNode)
		},
	}

	var startCmd = &cobra.Command{
		Use:   "start <node-file>",
		Short: "Starts the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.Start(currentNode)
		},
	}

	var stopCmd = &cobra.Command{
		Use:   "stop <node-file>",
		Short: "Stops the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.Stop(currentNode)
		},
	}

	var upgradeCmd = &cobra.Command{
		Use:   "upgrade <node-file>",
		Short: "Removes the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.Upgrade(currentNode)
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status <node-file>",
		Short: "Gives information about the current status",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			output, err := plugin.Status(currentNode)
			if err != nil {
				return err
			}

			fmt.Println(output)
			return nil
		},
	}

	var metaInfoCmd = &cobra.Command{
		Use:   "meta",
		Short: "Shows meta information such as allowed parameters for this plugin",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(plugin.Meta())
		},
	}

	var removeConfigCmd = &cobra.Command{
		Use:   "remove-config <node-file>",
		Short: "Remove the node configuration files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.RemoveConfig(currentNode)
		},
	}

	var removeDataCmd = &cobra.Command{
		Use:   "remove-data <node-file>",
		Short: "Remove the node data",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.RemoveData(currentNode)
		},
	}

	var removeRuntimeCmd = &cobra.Command{
		Use:   "remove-runtime <node-file>",
		Short: "Remove everything related to the node itself but no data or configs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.RemoveRuntime(currentNode)
		},
	}

	rootCmd.AddCommand(
		createConfigurationsCmd,
		startCmd,
		statusCmd,
		stopCmd,
		upgradeCmd,
		metaInfoCmd,
		removeConfigCmd,
		removeDataCmd,
		removeRuntimeCmd,
	)

	if funk.Contains(plugin.Meta().Supported, SupportsTest) {
		var testCmd = &cobra.Command{
			Use:   "test <node-file>",
			Short: "Runs a test suite against the running node",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				currentNode, err := node.Load(args[0])
				if err != nil {
					return err
				}

				success, err := plugin.Test(currentNode)

				if err != nil {
					return err
				}

				if !success {
					return fmt.Errorf("tests failed") // this causes a non-zero exit code
				}

				return nil
			},
		}

		rootCmd.AddCommand(testCmd)
	}

	// Start it all
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
