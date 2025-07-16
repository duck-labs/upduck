package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/utils"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS management commands",
	Long:  `Manage DNS forwarding and routing configurations.`,
}

var dnsForwardCmd = &cobra.Command{
	Use:   "forward [domain] [server] [port]",
	Short: "Forward domain to server (tower command)",
	Long: `Create an Nginx configuration to forward a domain to a specific server's private IP and port.
The server parameter can be either a server name or IP address from your connections.`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		server := args[1]
		serverPort := args[2]

		// Load connections to resolve server name/ID to IP address
		connectionsConfig, err := utils.LoadConnectionsConfig()
		if err != nil {
			return fmt.Errorf("failed to load connections config: %w", err)
		}

		serverIP, err := utils.ResolveServerToIP(connectionsConfig, server)
		if err != nil {
			return fmt.Errorf("failed to resolve server '%s': %w", server, err)
		}

		fmt.Printf("Configuring DNS forwarding for %s -> %s:%s (via %s)\n", domain, serverIP, serverPort, server)

		if err := utils.CreateNginxConfig(domain, serverIP, serverPort); err != nil {
			return fmt.Errorf("failed to create Nginx config: %w", err)
		}

		if err := exec.Command("nginx", "-t").Run(); err != nil {
			return fmt.Errorf("nginx configuration test failed: %w", err)
		}

		if err := exec.Command("systemctl", "reload", "nginx").Run(); err != nil {
			return fmt.Errorf("failed to reload Nginx: %w", err)
		}

		fmt.Printf("✅ Successfully configured DNS forwarding for %s\n", domain)
		fmt.Printf("Domain %s will now forward to %s:%s\n", domain, serverIP, serverPort)

		return nil
	},
}

func init() {
	dnsCmd.AddCommand(dnsForwardCmd)
}
