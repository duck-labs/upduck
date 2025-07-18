package network

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/pkg/config"
	"github.com/duck-labs/upduck/pkg/network"
	"github.com/duck-labs/upduck/types"
)

func getCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new network (tower command)",
		Long:  `Create a new virtual network on the local tower. Returns a network ID that can be used for server connections.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			connectionsConfig, err := config.LoadConnectionsConfig()
			if err != nil {
				return fmt.Errorf("failed to load connections config: %w", err)
			}

			wgNetworkBlock, err := network.GetNextAvailableNetworkBlock(connectionsConfig)
			if err != nil {
				return fmt.Errorf("failed to get network block: %w", err)
			}

			newNetwork := types.Network{
				ID:      network.GenerateTimeOrderedID(),
				Address: wgNetworkBlock.String(),
				Peers:   []types.Peer{},
			}

			connectionsConfig.Networks = append(connectionsConfig.Networks, newNetwork)

			if err := config.SaveConnectionsConfig(connectionsConfig); err != nil {
				return fmt.Errorf("failed to save connections config: %w", err)
			}

			fmt.Printf("✅ Network created successfully!\n")
			fmt.Printf("Network ID: %s\n", newNetwork.ID)
			fmt.Printf("Network Address: %s\n", newNetwork.Address)
			fmt.Printf("\nUse this Network ID when connecting servers:\n")
			fmt.Printf("  upduck network connect <tower-address> %s\n", newNetwork.ID)

			return nil
		},
	}
}
