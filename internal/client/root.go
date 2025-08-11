package client

import "github.com/spf13/cobra"

// NewCommand creates the root CLI command and attaches subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "Gophkeeper - Secure password and secrets manager",
		Long: `Gophkeeper is a CLI client for securely storing and managing 
passwords, tokens, and other secrets via HTTP or gRPC backends.`,
	}
	return cmd
}
