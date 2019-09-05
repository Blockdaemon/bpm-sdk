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
	"context"
	"fmt"
	"os"

	"io/ioutil"
	"strings"
	"time"

	"github.com/Blockdaemon/bpm-sdk/pkg/docker"
	"github.com/Blockdaemon/bpm-sdk/pkg/node"
	"github.com/Blockdaemon/bpm-sdk/pkg/template"
	"github.com/spf13/cobra"
)

// Plugin describes and provides the functionality for a plugin
type Plugin struct {
	// The name of the plugin
	Name string
	// A short one-line description of the plugin
	Description string
	// The semantic version of the plugin. Please increment with every change to the plugin
	Version string
	// Function that creates the secrets for a node
	CreateSecrets func(currentNode node.Node) error
	// Function that creates the configuration for the blockchain client
	CreateConfigs func(currentNode node.Node) error
	// Function to start the node. This usually involves creating a Docker network and starting containers
	Start func(currentNode node.Node) error
	// Function to remove a running node. This usually involves removing Docker resources and deleting generated configuration files
	Remove func(currentNode node.Node, purge bool) error
	// Function to upgrade a node with a new plugin version
	Upgrade func(currentNode node.Node) error
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

// DefaultRemove removes all configuration files and containers, volumes, network based on naming conventions
//
// Container names and volume names for a particular node all start with "bd-<node-id>".
func DefaultRemove(currentNode node.Node, purge bool) error {
	client, err := docker.NewBasicManager()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Remove containers
	containerNames, err := client.ListContainerNames(ctx)
	if err != nil {
		return err
	}

	for _, containerName := range containerNames {
		if strings.HasPrefix(containerName, currentNode.DockerPrefix()) {
			if err = client.ContainerAbsent(ctx, containerName); err != nil {
				return err
			}

		}
	}

	// Remove network
	if err = client.NetworkAbsent(ctx, currentNode.DockerNetworkName()); err != nil {
		return err
	}

	if purge {
		// Remove volumes
		volumeNames, err := client.ListVolumeIDs(ctx)
		if err != nil {
			return err
		}

		for _, volumeName := range volumeNames {
			if strings.HasPrefix(volumeName, currentNode.DockerPrefix()) {
				if err = client.VolumeAbsent(ctx, volumeName); err != nil {
					return err
				}

			}
		}

		// Remove all configuration files
		dir, err := ioutil.ReadDir(currentNode.ConfigsDirectory())
		if err != nil {
			return err
		}
		for _, d := range dir {
			if err := template.ConfigFileAbsent(d.Name(), currentNode); err != nil {
				return err
			}
		}
	}

	return nil
}

// DefaultCreateSecrets does nothing except printing that it does nothing
func DefaultCreateSecrets(currentNode node.Node) error {
	fmt.Println("Nothing to do here, skipping create-secrets")
	return nil
}

// DefaultCreateConfigs does nothing except printing that it does nothing
func DefaultCreateConfigs(currentNode node.Node) error {
	fmt.Println("Nothing to do here, skipping create-configurations")
	return nil
}

// DefaultUpgrade does nothing except printing that it does nothing
func DefaultUpgrade(currentNode node.Node) error {
	fmt.Println("Nothing to do here, skipping upgrade")
	return nil
}
