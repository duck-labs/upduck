package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/utils"
)

var rootCmd = &cobra.Command{
	Use:   "upduck",
	Short: "UpDuck - Self-hosted datacenter management CLI",
	Long: `UpDuck is a Golang CLI that helps developers create on-premise datacenters 
for self-hosted applications using existing tools like WireGuard, K3s, and Nginx.`,
}

var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Network management commands",
	Long:  `Manage virtual networks for server connections.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	config, err := utils.LoadConfig()
	if err != nil {
		if err.Error() == "not configured" {
			rootCmd.AddCommand(installCmd)
			rootCmd.AddCommand(versionCmd)
			return

		} else {
			log.Fatalf("failed to load config file")
			return
		}
	}

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(networkCmd)
	networkCmd.AddCommand(connectionsCmd)

	if config.Type == "tower" {
		rootCmd.AddCommand(dnsCmd)
		networkCmd.AddCommand(allowCmd)
		networkCmd.AddCommand(networkCreateCmd)
	}

	if config.Type == "server" {
		networkCmd.AddCommand(connectCmd)

	}
}
