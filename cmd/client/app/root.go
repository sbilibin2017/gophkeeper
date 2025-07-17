package app

import (
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands"
	"github.com/spf13/cobra"
)

// NewRootCommand creates the root CLI command with description.
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper — a password manager for secure private data storage",
		Long: `GophKeeper is a client-server system for securely storing
and managing logins, passwords, bank cards, and other private information.

Available commands allow user registration, authentication,
working with various secret types (logins, text, binary data, bank cards),
as well as synchronizing data with the server.`,
	}

	commands.RegisterRegisterCommand(rootCmd)
	commands.RegisterLoginCommand(rootCmd)
	commands.RegisterLogoutCommand(rootCmd)

	commands.RegisterAddBankCardCommand(rootCmd)
	commands.RegisterAddBinaryCommand(rootCmd)
	commands.RegisterAddTextCommand(rootCmd)
	commands.RegisterAddUsernamePasswordCommand(rootCmd)

	commands.RegisterGetSecretCommand(rootCmd)
	commands.RegisterListSecretsCommand(rootCmd)
	commands.RegisterDeleteSecretCommand(rootCmd)

	commands.RegisterSyncCommand(rootCmd)

	return rootCmd
}
