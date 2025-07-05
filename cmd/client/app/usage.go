package app

import "github.com/spf13/cobra"

// newUsageCommand creates a cobra.Command that displays usage information of the application.
// When called, it outputs the help for the root command.
func newUsageCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "usage",
		Short: "Show usage information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().Help()
		},
	}
}
