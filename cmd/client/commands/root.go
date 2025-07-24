package commands

import "github.com/spf13/cobra"

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gophkeeper",
		Short: "GophKeeper is a secure personal data manager",
		Long:  "GophKeeper CLI lets you register, login, and logout users securely using TLS authentication.",
	}

	return rootCmd
}
