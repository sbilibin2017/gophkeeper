package app

import "github.com/spf13/cobra"

// NewUsageCommand returns a cobra.Command that shows usage info.
func newUsageCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "usage",
		Short: "Show usage information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().Help()
		},
	}
}
