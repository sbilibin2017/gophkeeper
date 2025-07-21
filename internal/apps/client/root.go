package app

import "github.com/spf13/cobra"

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
