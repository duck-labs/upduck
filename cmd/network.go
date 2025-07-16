package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/types"
	"github.com/duck-labs/upduck/utils"
)

var networkCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new network (tower command)",
	Long:  `Create a new virtual network on the local tower. Returns a network ID that can be used for server connections.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		connectionsConfig, err := utils.LoadConnectionsConfig()
		if err != nil {
			return fmt.Errorf("failed to load connections config: %w", err)
		}

		wgNetworkBlock, err := utils.GetNextAvailableNetworkBlock(connectionsConfig)
		if err != nil {
			return fmt.Errorf("failed to get network block: %w", err)
		}

		newNetwork := types.Network{
			ID:      utils.GenerateTimeOrderedID(),
			Address: wgNetworkBlock.String(),
			Peers:   []types.Peer{},
		}

		connectionsConfig.Networks = append(connectionsConfig.Networks, newNetwork)

		if err := utils.SaveConnectionsConfig(connectionsConfig); err != nil {
			return fmt.Errorf("failed to save connections config: %w", err)
		}

		fmt.Printf("✅ Network created successfully!\n")
		fmt.Printf("Network ID: %s\n", newNetwork.ID)
		fmt.Printf("Network Address: %s\n", newNetwork.Address)
		fmt.Printf("\nUse this Network ID when connecting servers:\n")
		fmt.Printf("  upduck connect <tower-address> %s\n", newNetwork.ID)

		return nil
	},
}
