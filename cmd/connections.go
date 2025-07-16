package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/utils"
)

var connectionsCmd = &cobra.Command{
	Use:   "connections",
	Short: "Show connections and public key information",
	Long:  `Display information about connected servers/towers and show the public key digest.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rsaConfig, err := utils.LoadRSAKeys()
		if err != nil {
			return fmt.Errorf("failed to load WireGuard config: %w", err)
		}

		connectionsConfig, err := utils.LoadConnectionsConfig()
		if err != nil {
			return fmt.Errorf("failed to load connections config: %w", err)
		}

		fmt.Println("=== UpDuck Node Information ===")
		fmt.Printf("Public Key Digest: %s\n", utils.GetPublicKeyDigest(rsaConfig.PublicKey))
		fmt.Println()

		if len(connectionsConfig.Networks) > 0 {
			fmt.Println("=== Networks ===")
			for i, net := range connectionsConfig.Networks {
				fmt.Printf("Network %d:\n", i+1)
				fmt.Printf("   ID: %s\n", net.ID)
				fmt.Printf("   Address: %s\n", net.Address)
				fmt.Printf("   Peers: %d\n", len(net.Peers))
				for j, peer := range net.Peers {
					fmt.Printf("   Peer %d:\n", j+1)
					fmt.Printf("      ID: %s\n", peer.ID)
					fmt.Printf("      Address: %s\n", peer.Address)
					if peer.Endpoint != "" {
						fmt.Printf("      Endpoint: %s\n", peer.Endpoint)
					}
				}
				fmt.Println()
			}
		} else {
			fmt.Println("No networks found.")
		}

		if len(connectionsConfig.AllowedKeys) > 0 {
			fmt.Println("=== Allowed Server Keys ===")
			for i, key := range connectionsConfig.AllowedKeys {
				fmt.Printf("%d. (digest: %s)\n", i+1, key)
			}
		}

		return nil
	},
}
