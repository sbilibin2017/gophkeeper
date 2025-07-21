package client

import "github.com/spf13/cobra"

// NewRootCommand creates and returns the root Cobra command for the GophKeeper client CLI.
//
// The root command provides a brief description of the application and serves as
// the entry point for subcommands like registration, authentication, and managing secrets.
//
// Example usage:
//
//	gophkeeper [command] [flags]
//
// Available commands include user registration, authentication, working with various secret types
// (logins, text, binary data, bank cards), and synchronizing data with the server.
func NewRootCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — a password manager for secure private data storage",
		Long: `GophKeeper is a client-server system for securely storing
and managing logins, passwords, bank cards, and other private information.

Available commands allow user registration, authentication,
working with various secret types (logins, text, binary data, bank cards),
as well as synchronizing data with the server.`,
	}

	return cmd
}
