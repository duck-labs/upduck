package cmd

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/cmd/dns"
	"github.com/duck-labs/upduck/cmd/install"
	"github.com/duck-labs/upduck/cmd/network"
	"github.com/duck-labs/upduck/cmd/server"
	"github.com/duck-labs/upduck/cmd/version"
	"github.com/duck-labs/upduck/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "upduck",
	Short: "UpDuck - Self-hosted datacenter management CLI",
	Long: `UpDuck is a Golang CLI that helps developers create on-premise datacenters 
for self-hosted applications using existing tools like WireGuard, K3s, and Nginx.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	nodeConfig, err := config.LoadNodeConfig()
	if err != nil {
		if err.Error() == "not configured" {
			rootCmd.AddCommand(install.GetInstallCommand())
			rootCmd.AddCommand(version.GetVersionCommand())
			return
		} else {
			log.Fatalf("failed to load config file")
			return
		}
	}

	rootCmd.AddCommand(install.GetReinstallCommand())
	rootCmd.AddCommand(server.GetServerCommand())
	rootCmd.AddCommand(version.GetVersionCommand())

	networkCmd := network.GetNetworkCommand()

	if nodeConfig.Type == "tower" {
		rootCmd.AddCommand(dns.GetDNSCommand())
	}

	rootCmd.AddCommand(networkCmd)
}
