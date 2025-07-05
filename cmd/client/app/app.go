package app

import "github.com/spf13/cobra"

// NewAppCommand creates the root command "gophkeeper" and adds child commands to it.
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a CLI tool for securely managing personal data",
	}

	// Add child commands
	cmd.AddCommand(newBuildInfoCommand())
	cmd.AddCommand(newUsageCommand())
	cmd.AddCommand(newRegisterCommand())

	return cmd
}
