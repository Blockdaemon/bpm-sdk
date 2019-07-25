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
//			"gitlab.com/Blockdaemon/blockchain/bpm-lib/pkg/plugin"
//			"gitlab.com/Blockdaemon/blockchain/bpm-lib/pkg/node"
//	
//			"fmt"
//		)
//	
//		var pluginVersion string
//	
//		func createSecrets(currentNode node.Node) error {
//			fmt.Println("Nothing to do here, skipping create-secrets")
//			return nil
//		}
//	
//		func createConfigs(currentNode node.Node) error {
//			fmt.Println("Nothing to do here, skipping create-configurations")
//			return nil
//		}
//	
//		func start(currentNode node.Node) error {
//			fmt.Println("Nothing to do here, skipping start")
//			return nil
//		}
//	
//	
//		func remove(currentNode node.Node, purge bool) error {
//			fmt.Println("Nothing to do here, skipping remove")
//			return nil
//		}
//	
//		func upgrade(currentNode node.Node) error {
//			fmt.Println("Nothing to do here, skipping upgrade")
//			return nil
//		}
//	
//		func main() {
//			plugin.Initialize(plugin.Plugin{
//				Name: "empty",
//				Description: "A plugin that does nothing",
//				Version: pluginVersion,
//				CreateSecrets: createSecrets,
//				CreateConfigs: createConfigs,
//				Start: start,
//				Remove: remove,
//				Upgrade: upgrade,
//			})
//		}
package plugin

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gitlab.com/Blockdaemon/bpm-sdk/pkg/node"
)

// Plugin describes and provides the functionality for a plugin
type Plugin struct {
	// The name of the plugin
	Name          string
	// A short one-line description of the plugin
	Description   string
	// The semantic version of the plugin. Please increment with every change to the plugin
	Version       string
	// Function that creates the secrets for a node
	CreateSecrets func(currentNode node.Node) error
	// Function that creates the configuration for the blockchain client
	CreateConfigs func(currentNode node.Node) error
	// Function to start the node. This usually involves creating a Docker network and starting containers
	Start         func(currentNode node.Node) error
	// Function to remove a running node. This usually involves removing Docker resources and deleting generated configuration files
	Remove        func(currentNode node.Node, purge bool) error
	// Function to upgrade a node with a new plugin version
	Upgrade       func(currentNode node.Node) error
}

// Initialize creates the CLI for a plugin
func Initialize(plugin Plugin) {
	var baseDir string
	var purge bool

	// Initialize root command
	var rootCmd = &cobra.Command{
		Use:   plugin.Name,
		Short: plugin.Description,
	}

	pf := rootCmd.PersistentFlags()
	pf.StringVar(&baseDir, "base-dir", "~/.blockdaemon/", "The directory in which plugins and configuration is stored")

	// Create the commands
	var createSecretsCmd = &cobra.Command{
		Use:   "create-secrets <node-id>",
		Short: "Creates the secrets for a blockchain node and stores them on disk",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.LoadNode(baseDir, args[0])
			if err != nil {
				return err
			}

			return plugin.CreateSecrets(currentNode)
		},
	}
	rootCmd.AddCommand(createSecretsCmd)

	var createConfigurationsCmd = &cobra.Command{
		Use:   "create-configurations <node-id>",
		Short: "Creates the configurations for a blockchain node and stores them on disk",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.LoadNode(baseDir, args[0])
			if err != nil {
				return err
			}

			return plugin.CreateConfigs(currentNode)
		},
	}
	rootCmd.AddCommand(createConfigurationsCmd)

	var startCmd = &cobra.Command{
		Use:   "start <node-id>",
		Short: "Starts the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
				currentNode, err := node.LoadNode(baseDir, args[0])
				if err != nil {
					return err
				}

				return plugin.Start(currentNode)
		},
	}
	rootCmd.AddCommand(startCmd)

	var removeCmd = &cobra.Command{
		Use:   "remove <node-id>",
		Short: "Removes the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.LoadNode(baseDir, args[0])
			if err != nil {
				return err
			}

			return plugin.Remove(currentNode, purge)
		},
	}
	removeCmd.Flags().BoolVar(&purge, "purge", false, "Purge all data, secrets and configuration files")
	rootCmd.AddCommand(removeCmd)

	var upgrade = &cobra.Command{
		Use:   "upgrade <node-id>",
		Short: "Removes the docker containers",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			currentNode, err := node.LoadNode(baseDir, args[0])
			if err != nil {
				return err
			}

			return plugin.Upgrade(currentNode)
		},
	}
	rootCmd.AddCommand(upgrade)

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version of this plugin",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(plugin.Version)
		},
	}
	rootCmd.AddCommand(versionCmd)

	// Start it all
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
