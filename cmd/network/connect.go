package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/pkg/config"
	"github.com/duck-labs/upduck/pkg/crypto"
	"github.com/duck-labs/upduck/pkg/types"
)

func getConnectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "connect <tower-dns> <network-id>",
		Short: "Connect to a tower (server command)",
		Long:  `Connect this server to a tower node. The tower must have allowed this server's public key first.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			towerAddress := args[0]
			networkID := args[1]

			wgConfig, err := config.LoadWireguardConfig()
			if err != nil {
				return fmt.Errorf("failed to load WireGuard config: %w", err)
			}

			rsaConfig, err := crypto.LoadRSAKeys()
			if err != nil {
				return fmt.Errorf("failed to load RSA config: %w", err)
			}

			request := types.ConnectRequest{
				PublicKey:   rsaConfig.PublicKey,
				WGPublicKey: wgConfig.PublicKey,
			}

			requestData, err := json.Marshal(request)
			if err != nil {
				return fmt.Errorf("failed to marshal request: %w", err)
			}

			towerURL, err := url.Parse(towerAddress)
			if err != nil {
				return fmt.Errorf("failed to parse tower URL: %w", err)
			}

			apiURL := fmt.Sprintf("%s/api/servers/network/%s/connect", towerAddress, networkID)
			fmt.Printf("Connecting to tower at %s...\n", apiURL)

			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(requestData))
			if err != nil {
				return fmt.Errorf("failed to connect to tower: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusConflict {
				return fmt.Errorf("conflict: server already connected to this tower")
			}

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("tower responded with status %d", resp.StatusCode)
			}

			var response types.ConnectResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}

			connectionsConfig, err := config.LoadConnectionsConfig()
			if err != nil {
				return fmt.Errorf("failed to load connections config: %w", err)
			}

			peer := types.Peer{
				ID:        response.PeerID,
				PublicKey: response.WGPublicKey,
				Address:   response.WGNetworkBlock,
				Endpoint:  towerURL.Hostname(),
			}

			var existingNetworkIndex = -1
			for i, network := range connectionsConfig.Networks {
				if network.ID == response.NetworkID {
					existingNetworkIndex = i
					break
				}
			}

			if existingNetworkIndex >= 0 {
				connectionsConfig.Networks[existingNetworkIndex].Peers = append(connectionsConfig.Networks[existingNetworkIndex].Peers, peer)
			} else {
				network := types.Network{
					ID:      response.NetworkID,
					Address: response.WGAddress,
					Peers:   []types.Peer{peer},
				}
				connectionsConfig.Networks = append(connectionsConfig.Networks, network)
			}

			if err := config.SaveConnectionsConfig(connectionsConfig); err != nil {
				return fmt.Errorf("failed to save connections config: %w", err)
			}

			fmt.Printf("✅ Successfully connected to tower %s\n", towerAddress)
			fmt.Printf("Network block: %s\n", response.WGNetworkBlock)
			fmt.Printf("This node address: %s\n", response.WGAddress)

			return nil
		},
	}
}
