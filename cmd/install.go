package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/duck-labs/upduck-v2/server"
	"github.com/duck-labs/upduck-v2/utils"
)

var installCmd = &cobra.Command{
	Use:   "install [server|tower]",
	Short: "Install and configure upduck as server or tower",
	Long: `Install and configure upduck as either a server or tower node.
This command will:
- Install WireGuard and generate keys
- Install additional dependencies (K3s for server, Nginx for tower)
- Start the upduck HTTP server as a systemd service`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nodeType := args[0]
		if nodeType != "server" && nodeType != "tower" {
			return fmt.Errorf("invalid node type: %s (must be 'server' or 'tower')", nodeType)
		}

		fmt.Printf("Installing upduck as %s...\n", nodeType)

		if os.Geteuid() != 0 {
			return fmt.Errorf("this command must be run as root")
		}

		if !utils.IsWireguardInstalled() {
			if err := utils.InstallWireguard(); err != nil {
				return fmt.Errorf("failed to install WireGuard: %w", err)
			}
		} else {
			fmt.Println("WireGuard is already installed")
		}

		if _, err := utils.LoadWireguardConfig(); err != nil {
			fmt.Println("Generating WireGuard keys...")
			wgConfig, err := utils.GenerateWireguardKeys()
			if err != nil {
				return fmt.Errorf("failed to generate WireGuard keys: %w", err)
			}

			if err := utils.SaveWireguardConfig(wgConfig); err != nil {
				return fmt.Errorf("failed to save WireGuard config: %w", err)
			}
			fmt.Printf("WireGuard keys generated and saved to %s\n", utils.WireguardConfigFile)
		} else {
			fmt.Println("WireGuard keys already exist")
		}

		switch nodeType {
		case "server":
			if !utils.IsK3sInstalled() {
				if err := utils.InstallK3s(); err != nil {
					return fmt.Errorf("failed to install K3s: %w", err)
				}
			} else {
				fmt.Println("K3s is already installed")
			}
		case "tower":
			if !utils.IsNginxInstalled() {
				if err := utils.InstallNginx(); err != nil {
					return fmt.Errorf("failed to install Nginx: %w", err)
				}
			} else {
				fmt.Println("Nginx is already installed")
			}
		}

		if err := createSystemdService(nodeType); err != nil {
			return fmt.Errorf("failed to create systemd service: %w", err)
		}

		fmt.Printf("✅ Successfully installed upduck as %s\n", nodeType)
		fmt.Println("The upduck HTTP server is now running as a systemd service")
		
		return nil
	},
}

func createSystemdService(nodeType string) error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=UpDuck %s Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=%s server --type=%s
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, nodeType, execPath, nodeType)

	serviceFile := "/etc/systemd/system/upduck.service"
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return err
	}

	if err := server.RunCommand("systemctl", "daemon-reload"); err != nil {
		return err
	}

	if err := server.RunCommand("systemctl", "enable", "upduck"); err != nil {
		return err
	}

	if err := server.RunCommand("systemctl", "start", "upduck"); err != nil {
		return err
	}

	return nil
}
