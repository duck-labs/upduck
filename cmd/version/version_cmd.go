package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of UpDuck",
		Long:  `Show version information for UpDuck CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("UpDuck %s\n", version)
		},
	}
}
