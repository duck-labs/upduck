package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/duck-labs/upduck-v2/utils"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "DNS management commands",
	Long:  `Manage DNS forwarding and routing configurations.`,
}

var dnsForwardCmd = &cobra.Command{
	Use:   "forward [domain] [server] [server-local-address:port]",
	Short: "Forward domain to server (tower command)",
	Long: `Create an Nginx configuration to forward a domain to a specific server's private IP and port.
The server parameter can be either a server name or IP address from your connections.`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := args[0]
		server := args[1]
		addressPort := args[2]

		parts := strings.Split(addressPort, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid address:port format: %s (expected format: IP:PORT)", addressPort)
		}
		serverIP := parts[0]
		port := parts[1]

		fmt.Printf("Configuring DNS forwarding for %s -> %s:%s (via %s)\n", domain, serverIP, port, server)

		if err := utils.CreateNginxConfig(domain, serverIP, port); err != nil {
			return fmt.Errorf("failed to create Nginx config: %w", err)
		}

		if err := exec.Command("nginx", "-t").Run(); err != nil {
			return fmt.Errorf("Nginx configuration test failed: %w", err)
		}

		if err := exec.Command("systemctl", "reload", "nginx").Run(); err != nil {
			return fmt.Errorf("failed to reload Nginx: %w", err)
		}

		fmt.Printf("✅ Successfully configured DNS forwarding for %s\n", domain)
		fmt.Printf("Domain %s will now forward to %s:%s\n", domain, serverIP, port)

		return nil
	},
}

func init() {
	dnsCmd.AddCommand(dnsForwardCmd)
}
