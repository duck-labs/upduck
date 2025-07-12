package cmd

import (
	"github.com/spf13/cobra"
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
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(connectionsCmd)
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(allowCmd)
	rootCmd.AddCommand(dnsCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(versionCmd)
}
