package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck-v2/utils"
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

		if len(connectionsConfig.Connections) > 0 {
			fmt.Println("=== Connected Nodes ===")
			for i, conn := range connectionsConfig.Connections {
				fmt.Printf("%d. Type: %s\n", i+1, conn.Type)
				if conn.DNS != "" {
					fmt.Printf("   DNS: %s\n", conn.DNS)
				}
				fmt.Printf("   Public Key: %s\n", conn.PublicKey)

				if conn.WGAddress != "" {
					fmt.Printf("   WG Address: %s\n", conn.WGAddress)
				}

				if conn.WGNetworkBlock != "" {
					fmt.Printf("   WG Network Block: %s\n", conn.WGNetworkBlock)
				}

				fmt.Println()
			}
		} else {
			fmt.Println("No connections found.")
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
