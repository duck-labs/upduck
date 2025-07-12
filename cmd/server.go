package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/duck-labs/upduck-v2/server"
)

var (
	serverNodeType string
	serverPort     string
)

var serverCmd = &cobra.Command{
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

		srv := server.NewServer(serverNodeType, serverPort)
		return srv.Start()
	},
}

func init() {
	serverCmd.Flags().StringVar(&serverNodeType, "type", "", "Node type (server or tower)")
	serverCmd.Flags().StringVar(&serverPort, "port", "8080", "Port to listen on")
	serverCmd.MarkFlagRequired("type")
}
