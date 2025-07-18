package network

import (
	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/pkg/config"
)

func GetNetworkCommand() *cobra.Command {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Network management commands",
		Long:  `Manage virtual networks for server connections.`,
	}

	networkCmd.AddCommand(getConnectionsCommand())

	nodeConfig, err := config.LoadNodeConfig()
	if err == nil {
		if nodeConfig.Type == "tower" {
			networkCmd.AddCommand(getCreateCommand())
			networkCmd.AddCommand(getAllowCommand())
		}

		if nodeConfig.Type == "server" {
			networkCmd.AddCommand(getConnectCommand())
		}
	}

	return networkCmd
}
