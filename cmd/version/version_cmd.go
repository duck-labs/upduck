package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

func getVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of UpDuck",
		Long:  `Show version information for UpDuck CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("UpDuck v1.0.0")
		},
	}
}
