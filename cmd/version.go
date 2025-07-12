package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version, commit, and build date information for upduck.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("upduck version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
	},
}
