// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
//
// Please see the main BPM-SDK documentation for more details on how to implement a new plugin.
package plugin

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thoas/go-funk"
	"go.blockdaemon.com/bpm/sdk/pkg/node"
)

// ParameterValidator provides a function to validate the node parameters
type ParameterValidator interface {
	// ValidateParameters validates the ndoe parameters
	ValidateParameters(currentNode node.Node) error
}

// IdentityCreator provides functions to create and remove the identity (e.g. private keys) of a node
type IdentityCreator interface {
	// Function that creates the identity of a node
	CreateIdentity(currentNode node.Node) error

	// Removes identity related to the node
	RemoveIdentity(currentNode node.Node) error
}

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

	ParameterValidator
	IdentityCreator
	Configurator
	LifecycleHandler
	Upgrader
	Tester
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
	var validateParametersCmd = &cobra.Command{
		Use:   "validate-parameters <node-file>",
		Short: "Validates the parameters in the node file",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.ValidateParameters(currentNode)
		},
	}

	var createConfigurationsCmd = &cobra.Command{
		Use:   "create-configurations <node-file>",
		Short: "Creates the configurations for a node",
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
		Short: "Starts the node",
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
		Short: "Stops the node",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(args[0])
			if err != nil {
				return err
			}

			return plugin.Stop(currentNode)
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status <node-file>",
		Short: "Gives information about the current node status",
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
		Short: "Shows meta information for this package",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(plugin.Meta())
		},
	}

	var removeConfigCmd = &cobra.Command{
		Use:   "remove-config <node-file>",
		Short: "Removes the node configuration",
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
		Short: "Removes the node data (i.e. already synced blockchain)",
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
		Short: "Removes everything related to the node itself but no data, identity or configs",
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
		validateParametersCmd,
		createConfigurationsCmd,
		startCmd,
		statusCmd,
		stopCmd,
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

	if funk.Contains(plugin.Meta().Supported, SupportsUpgrade) {
		var upgradeCmd = &cobra.Command{
			Use:   "upgrade <node-file>",
			Short: "Upgrades the node to a newer version of a package",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				currentNode, err := node.Load(args[0])
				if err != nil {
					return err
				}

				return plugin.Upgrade(currentNode)
			},
		}

		rootCmd.AddCommand(upgradeCmd)
	}

	if funk.Contains(plugin.Meta().Supported, SupportsIdentity) {
		var createIdentityCmd = &cobra.Command{
			Use:   "create-identity <node-file>",
			Short: "Creates the nodes identity (e.g. private keys, certificates, etc.)",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				currentNode, err := node.Load(args[0])
				if err != nil {
					return err
				}

				return plugin.CreateIdentity(currentNode)
			},
		}

		var removeIdentityCmd = &cobra.Command{
			Use:   "remove-identity <node-file>",
			Short: "Removes the node identity",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				currentNode, err := node.Load(args[0])
				if err != nil {
					return err
				}

				return plugin.RemoveIdentity(currentNode)
			},
		}

		rootCmd.AddCommand(
			createIdentityCmd,
			removeIdentityCmd,
		)
	}

	// Start it all
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
