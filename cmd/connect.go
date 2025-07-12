package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck-v2/types"
	"github.com/duck-labs/upduck-v2/utils"
)

var connectCmd = &cobra.Command{
	Use:   "connect [tower-dns]",
	Short: "Connect to a tower (server command)",
	Long:  `Connect this server to a tower node. The tower must have allowed this server's public key first.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		towerAddress := args[0]

		wgConfig, err := utils.LoadWireguardConfig()
		if err != nil {
			return fmt.Errorf("failed to load WireGuard config: %w", err)
		}

		rsaConfig, err := utils.LoadRSAKeys()
		if err != nil {
			return fmt.Errorf("failed to load WireGuard config: %w", err)
		}

		request := types.ConnectRequest{
			PublicKey:   rsaConfig.PublicKey,
			WGPublicKey: wgConfig.PublicKey,
		}

		requestData, err := json.Marshal(request)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		url := fmt.Sprintf("http://%s/api/servers/connect", towerAddress)
		fmt.Printf("Connecting to tower at %s...\n", url)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestData))
		if err != nil {
			return fmt.Errorf("failed to connect to tower: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("tower responded with status %d", resp.StatusCode)
		}

		var response types.ConnectResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		connectionsConfig, err := utils.LoadConnectionsConfig()
		if err != nil {
			return fmt.Errorf("failed to load connections config: %w", err)
		}

		peer := types.Peer{
			PublicKey: response.WGPublicKey,
			Address:   response.WGNetworkBlock,
			Endpoint:  towerAddress,
		}

		network := types.Network{
			Address: response.WGAddress,
			Peers:   []types.Peer{peer},
		}

		connectionsConfig.Networks = append(connectionsConfig.Networks, network)

		if err := utils.SaveConnectionsConfig(connectionsConfig); err != nil {
			return fmt.Errorf("failed to save connections config: %w", err)
		}

		fmt.Printf("✅ Successfully connected to tower %s\n", towerAddress)
		fmt.Printf("Network block: %s\n", response.WGNetworkBlock)
		fmt.Printf("This node address: %s\n", response.WGAddress)

		return nil
	},
}
