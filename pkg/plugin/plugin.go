// Package plugin provides an easy way to create the required CLI for a plugin.
// It abstracts away all the command line and file parsing so users just need to implement the actual logic.
//
// Please see the main BPM-SDK documentation for more details on how to implement a new plugin.
//
// Usage Example:
//
//		package main
//
//		import (
//			"github.com/Blockdaemon/bpm-sdk/pkg/plugin"
//			"github.com/Blockdaemon/bpm-sdk/pkg/node"
//
//			"fmt"
//		)
//
//		var pluginVersion string
//
//		func start(currentNode node.Node) error {
//			fmt.Println("Nothing to do here, skipping start")
//			return nil
//		}
//
//		func main() {
//			plugin.Initialize(plugin.Plugin{
//				Name: "empty",
//				Description: "A plugin that does nothing",
//				Version: pluginVersion,
//				Start: start,
//				CreateSecrets: plugin.DefaultCreateSecrets,
//				CreateConfigs: plugin.DefaultCreateConfigs
//				Remove: plugin.DefaultRemove,
//				Upgrade: plugin.DefaultUpgrade,
//			})
//		}
package plugin

import (
	"fmt"
	"os"

	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/spf13/cobra"
)

// Plugin describes and provides the functionality for a plugin
type Plugin interface {
	// Returns the name of the plugin
	Name() string
	// Returns a short one-line description of the plugin
	Description() string
	// Returns the semantic version of the plugin. Please increment with every change to the plugin
	Version() string
	// Function that creates the secrets for a node
	CreateSecrets(currentNode node.Node) error
	// Function that creates the configuration for the blockchain client
	CreateConfigs(currentNode node.Node) error
	// Function to start the node. This usually involves creating a Docker network and starting containers
	Start(currentNode node.Node) error
	// Function to stop a running node. This usually involves removing Docker containers
	Stop(currentNode node.Node, purge bool) error
	// Function to return the status (running, incomplete, stopped) of a  node
	Status(currentNode node.Node) (string, error)
	// Function to upgrade a node with a new plugin version
	Upgrade(currentNode node.Node) error
	// Function to run tests against the node
	Test(currentNode node.Node) (bool, error)
	// Returns available parameters to configure a node
	Parameters() string
}

// Initialize creates the CLI for a plugin
func Initialize(plugin Plugin) {
	var baseDir string
	var purge bool

	// Initialize root command
	var rootCmd = &cobra.Command{
		Use:          plugin.Name(),
		Short:        plugin.Description(),
		SilenceUsage: true,
	}

	pf := rootCmd.PersistentFlags()
	pf.StringVar(&baseDir, "base-dir", "~/.bpm/nodes", "The directory in which the node secrets and configuration are stored")

	// Create the commands
	var createSecretsCmd = &cobra.Command{
		Use:   "create-secrets <node-id>",
		Short: "Creates the secrets for a blockchain node and stores them on disk",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(baseDir, args[0])

			if err != nil {
				return err
			}

			return plugin.CreateSecrets(currentNode)
		},
	}

	var createConfigurationsCmd = &cobra.Command{
		Use:   "create-configurations <node-id>",
		Short: "Creates the configurations for a blockchain node and stores them on disk",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(baseDir, args[0])
			if err != nil {
				return err
			}

			return plugin.CreateConfigs(currentNode)
		},
	}

	var startCmd = &cobra.Command{
		Use:   "start <node-id>",
		Short: "Starts the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(baseDir, args[0])
			if err != nil {
				return err
			}

			return plugin.Start(currentNode)
		},
	}

	var stopCmd = &cobra.Command{
		Use:   "stop <node-id>",
		Short: "Stops the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(baseDir, args[0])
			if err != nil {
				return err
			}

			return plugin.Stop(currentNode, purge)
		},
	}
	stopCmd.Flags().BoolVar(&purge, "purge", false, "Purge all data volumes and configuration files")

	var upgradeCmd = &cobra.Command{
		Use:   "upgrade <node-id>",
		Short: "Removes the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(baseDir, args[0])
			if err != nil {
				return err
			}

			return plugin.Upgrade(currentNode)
		},
	}

	var statusCmd = &cobra.Command{
		Use:   "status <node-id>",
		Short: "Gives information about the current status",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(baseDir, args[0])
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

	var testCmd = &cobra.Command{
		Use:   "test <node-id>",
		Short: "Runs a test suite against the running node",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.Load(baseDir, args[0])
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

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version of this plugin",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(plugin.Version)
		},
	}

	var parametersCmd = &cobra.Command{
		Use:   "parameters",
		Short: "Shows allowed parameters for node.json",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(plugin.Parameters())
		},
	}

	rootCmd.AddCommand(
		createSecretsCmd,
		createConfigurationsCmd,
		startCmd,
		statusCmd,
		testCmd,
		stopCmd,
		upgradeCmd,
		versionCmd,
		parametersCmd,
	)

	// Start it all
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
