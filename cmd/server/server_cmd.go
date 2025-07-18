package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/pkg/api"
	"github.com/duck-labs/upduck/pkg/config"
)

var (
	serverPort string
)

func getServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "server",
		Short:  "Start the upduck HTTP server",
		Long:   `Start the upduck HTTP server that handles API requests.`,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeConfig, err := config.LoadNodeConfig()
			if err != nil {
				log.Fatalf("failed to load config file")
			}

			if nodeConfig.Type != "server" && nodeConfig.Type != "tower" {
				return fmt.Errorf("invalid node type: %s (must be 'server' or 'tower')", nodeConfig.Type)
			}

			srv := api.NewServer(nodeConfig.Type, serverPort)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				if err := srv.Start(); err != nil {
					fmt.Printf("Server error: %v\n", err)
					os.Exit(1)
				}
			}()

			<-sigChan
			srv.Stop()

			return nil
		},
	}

	cmd.Flags().StringVar(&serverPort, "port", "8080", "Port to listen on")
	cmd.MarkFlagRequired("type")

	return cmd
}
