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
			for _, net := range connectionsConfig.Networks {
				fmt.Printf("   Peers: %d\n", len(net.Peers))
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
