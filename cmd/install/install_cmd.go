package install

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/pkg/config"
	"github.com/duck-labs/upduck/pkg/crypto"
	"github.com/duck-labs/upduck/pkg/network"
	"github.com/duck-labs/upduck/pkg/system"
)

func getInstallCommand() *cobra.Command {
	return &cobra.Command{
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
				return fmt.Errorf("this command must be run as root (use sudo)")
			}

			if !system.IsWireguardInstalled() {
				if err := system.InstallWireguard(); err != nil {
					return fmt.Errorf("failed to install WireGuard: %w", err)
				}
			} else {
				fmt.Println("WireGuard is already installed")
			}

			_, err := config.LoadWireguardConfig()
			if err != nil {
				fmt.Println("Generating WireGuard keys...")
				wgConfig, err := network.GenerateWireguardKeys()
				if err != nil {
					return fmt.Errorf("failed to generate WireGuard keys: %w", err)
				}

				if err := config.SaveWireguardConfig(wgConfig); err != nil {
					return fmt.Errorf("failed to save WireGuard config: %w", err)
				}
				fmt.Printf("WireGuard keys generated and saved to [%s]\n", config.WireguardConfigFile)
			} else {
				fmt.Println("WireGuard keys already exist")
			}

			_, err = crypto.LoadRSAKeys()
			if err != nil {
				fmt.Println("Generating RSA keys...")
				_, err := crypto.GenerateRSAKeys()
				if err != nil {
					return fmt.Errorf("failed to generate RSA keys: %w", err)
				}
				fmt.Printf("RSA keys generated and saved to [%s] and [%s]\n", config.RSAPrivateKey, config.RSAPublicKey)
			} else {
				fmt.Println("RSA keys already exist")
			}

			err = config.WriteNodeConfig(nodeType)
			if err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			switch nodeType {
			case "server":
				if !system.IsK3sInstalled() {
					if err := system.InstallK3s(); err != nil {
						return fmt.Errorf("failed to install K3s: %w", err)
					}
				} else {
					fmt.Println("K3s is already installed")
				}
			case "tower":
				if !system.IsNginxInstalled() {
					if err := system.InstallNginx(); err != nil {
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
ExecStart=%s server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
`, nodeType, execPath)

	serviceFile := "/etc/systemd/system/upduck.service"
	if err := os.WriteFile(serviceFile, []byte(serviceContent), 0644); err != nil {
		return err
	}

	if err := system.RunCommand("systemctl", "daemon-reload"); err != nil {
		return err
	}

	if err := system.RunCommand("systemctl", "enable", "upduck"); err != nil {
		return err
	}

	if err := system.RunCommand("systemctl", "start", "upduck"); err != nil {
		return err
	}

	return nil
}
