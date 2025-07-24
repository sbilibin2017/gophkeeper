package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/cmd/client/commands"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	rootCmd := commands.NewRootCommand()

	// Auth & session commands
	rootCmd.AddCommand(commands.NewRegisterCommand())
	rootCmd.AddCommand(commands.NewLoginCommand())
	rootCmd.AddCommand(commands.NewLogoutCommand())

	// Add secret commands
	rootCmd.AddCommand(commands.NewAddBankCardCommand())
	rootCmd.AddCommand(commands.NewAddBinaryCommand())
	rootCmd.AddCommand(commands.NewAddTextCommand())
	rootCmd.AddCommand(commands.NewAddUserCommand())

	// Read commands
	rootCmd.AddCommand(commands.NewGetCommand())
	rootCmd.AddCommand(commands.NewListCommand())

	// Sync command
	rootCmd.AddCommand(commands.NewSyncCommand())

	return rootCmd.Execute()
}
