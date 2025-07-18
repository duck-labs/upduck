package dns

import (
	"github.com/spf13/cobra"
)

func GetDNSCommand() *cobra.Command {
	dnsCmd := &cobra.Command{
		Use:   "dns",
		Short: "DNS management commands",
		Long:  `Manage DNS forwarding and routing configurations.`,
	}

	dnsCmd.AddCommand(getForwardCommand())

	return dnsCmd
}
