package version

import (
	"github.com/spf13/cobra"
)

func GetVersionCommand() *cobra.Command {
	return getVersionCommand()
}
