// Package main is the entry point for the GophKeeper CLI application.
package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/inernal/apps/client"
)

// main initializes and runs the GophKeeper CLI.
// It logs any fatal errors that occur during execution.
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

// run configures the root Cobra command and registers all available subcommands.
// It returns an error if command execution fails.
func run() error {
	rootCmd := client.NewCommand()

	// Auth
	rootCmd.AddCommand(client.NewRegisterCommand())
	rootCmd.AddCommand(client.NewLoginCommand())

	// Add secret
	rootCmd.AddCommand(client.NewAddBankCardCommand())
	rootCmd.AddCommand(client.NewAddBinaryCommand())
	rootCmd.AddCommand(client.NewAddTextCommand())
	rootCmd.AddCommand(client.NewAddUserCommand())

	// List secrets
	rootCmd.AddCommand(client.NewListCommand())

	// Sync secrets with server
	rootCmd.AddCommand(client.NewSyncCommand())

	return rootCmd.Execute()
}
