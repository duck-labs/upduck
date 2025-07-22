package install

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/pkg/config"
	"github.com/duck-labs/upduck/pkg/system"
)

func GetReinstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reinstall",
		Short: "Recreate the systemd service",
		Long:  "Recreate the systemd service for upduck based on the current node configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Recreating upduck systemd service...")

			if os.Geteuid() != 0 {
				return fmt.Errorf("this command must be run as root (use sudo)")
			}

			nodeConfig, err := config.LoadNodeConfig()
			if err != nil {
				return fmt.Errorf("failed to load node configuration: %w", err)
			}

			if nodeConfig.Type == "" {
				return fmt.Errorf("node type not configured - run 'upduck install [server|tower]' first")
			}

			fmt.Println("Stopping existing service...")
			system.RunCommand("systemctl", "stop", "upduck")

			if err := createSystemdService(nodeConfig.Type); err != nil {
				return fmt.Errorf("failed to recreate systemd service: %w", err)
			}

			fmt.Printf("✅ Successfully recreated systemd service for %s node\n", nodeConfig.Type)
			fmt.Println("The upduck HTTP server is now running as a systemd service")

			return nil
		},
	}
}
