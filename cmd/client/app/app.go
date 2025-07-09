package app

import (
	"github.com/spf13/cobra"
)

// NewAppCommand creates the root command for the GophKeeper CLI application.
// It includes all available subcommands for managing private data.
func NewAppCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — CLI manager for private data",
		Long:  "GophKeeper is a CLI tool for securely managing your private data such as logins, texts, files, cards, and more.",
	}

	// Add subcommands for various functionalities
	cmd.AddCommand(newVersionCommand())          // Shows version and build info
	cmd.AddCommand(newConfigCommand())           // Configure client parameters (token, server URL)
	cmd.AddCommand(newRegisterCommand())         // Register a new user account
	cmd.AddCommand(newLoginCommand())            // Login existing user
	cmd.AddCommand(newAddLoginPasswordCommand()) // Add a login and password secret
	cmd.AddCommand(newAddTextCommand())          // Add arbitrary text secret
	cmd.AddCommand(newAddBinaryCommand())        // Add binary data secret from a file
	cmd.AddCommand(newAddCardCommand())          // Add bank card secret
	cmd.AddCommand(newListCommand())             // List saved secrets
	cmd.AddCommand(newSyncCommand())             // Synchronize local data with server

	return cmd
}
