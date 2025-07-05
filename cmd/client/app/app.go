package app

import "github.com/spf13/cobra"

// NewAppCommand creates the root command "gophkeeper" and adds subcommands to it.
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure personal data manager CLI",
	}

	// Add subcommands
	cmd.AddCommand(newBuildInfoCommand())
	cmd.AddCommand(newUsageCommand())
	cmd.AddCommand(newRegisterCommand())

	return cmd
}
