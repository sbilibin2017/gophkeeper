package commands

import "github.com/spf13/cobra"

// NewRootCommand creates the root command for the GophKeeper CLI application.
//
// This command serves as the entry point for all subcommands related to
// managing secure personal data, such as user registration, login, and logout.
// It provides a brief description of the application and sets up the CLI hierarchy.
func NewRootCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure personal data manager",
		Long:  "GophKeeper CLI lets you register, login, and logout users securely using TLS authentication.",
	}
}
