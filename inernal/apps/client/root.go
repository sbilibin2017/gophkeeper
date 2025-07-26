package client

import (
	"github.com/spf13/cobra"
)

// NewRootCommand returns the root cobra command for the GophKeeper CLI.
// It serves as the entry point for all subcommands and provides basic usage information.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure personal data manager",
		Long:  "GophKeeper CLI lets you register, login, and logout users securely using TLS authentication.",
	}
}
