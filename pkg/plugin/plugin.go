package plugin

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gitlab.com/Blockdaemon/bpm-sdk/pkg/node"
)

type Plugin struct {
	Name          string
	Description   string
	Version       string
	CreateSecrets func(currentNode node.Node) error
	CreateConfigs func(currentNode node.Node) error
	Start         func(currentNode node.Node) error
	Remove        func(currentNode node.Node, purge bool) error
	Upgrade       func(currentNode node.Node) error
}

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
	upgrade.Flags().BoolVar(&purge, "purge", false, "Purge all data, secrets and configuration files")
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
