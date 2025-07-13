package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/utils"
)

var allowCmd = &cobra.Command{
	Use:   "allow [server-pub-key]",
	Short: "Allow a server to connect (tower command)",
	Long:  `Add a server's public key to the list of allowed servers that can connect to this tower.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverPubKeyDigest := args[0]

		connectionsConfig, err := utils.LoadConnectionsConfig()
		if err != nil {
			return fmt.Errorf("failed to load connections config: %w", err)
		}

		for _, allowedKey := range connectionsConfig.AllowedKeys {
			if allowedKey == serverPubKeyDigest {
				fmt.Printf("Server public key %s is already allowed\n", serverPubKeyDigest)
				return nil
			}
		}

		connectionsConfig.AllowedKeys = append(connectionsConfig.AllowedKeys, serverPubKeyDigest)

		if err := utils.SaveConnectionsConfig(connectionsConfig); err != nil {
			return fmt.Errorf("failed to save connections config: %w", err)
		}

		fmt.Printf("✅ Successfully allowed server with public key digest: %s\n", serverPubKeyDigest)

		return nil
	},
}
