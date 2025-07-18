package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck/pkg/api"
)

var (
	serverNodeType string
	serverPort     string
)

func getServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "server",
		Short:  "Start the upduck HTTP server",
		Long:   `Start the upduck HTTP server that handles API requests.`,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverNodeType == "" {
				return fmt.Errorf("node type must be specified with --type flag")
			}

			if serverNodeType != "server" && serverNodeType != "tower" {
				return fmt.Errorf("invalid node type: %s (must be 'server' or 'tower')", serverNodeType)
			}

			srv := api.NewServer(serverNodeType, serverPort)

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

	cmd.Flags().StringVar(&serverNodeType, "type", "", "Node type (server or tower)")
	cmd.Flags().StringVar(&serverPort, "port", "8080", "Port to listen on")
	cmd.MarkFlagRequired("type")

	return cmd
}
