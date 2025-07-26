// Package main is the entry point of the GophKeeper CLI application.
//
// GophKeeper is a secure personal data manager that allows users to store
// various secret types (bank cards, login credentials, binary data, etc.).
// This CLI interacts with a secure backend and handles operations such as registration,
// login, adding secrets, listing, and syncing.
//
// Usage example:
//
//	gophkeeper register --server-url https://api.example.com ...
package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/inernal/cli"
)

var (
	// buildVersion holds the version of the build.
	// It is intended to be injected at build time via -ldflags.
	buildVersion = "N/A"

	// buildDate holds the date of the build.
	// It is intended to be injected at build time via -ldflags.
	buildDate = "N/A"
)

// main is the entry point of the GophKeeper CLI.
// It initializes and executes the root command.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run sets up the root command and its subcommands,
// then executes the CLI logic.
//
// It returns an error if command execution fails.
func run() error {
	rootCmd := cli.NewClientCommand()

	rootCmd.AddCommand(cli.NewClientRegisterCommand())
	rootCmd.AddCommand(cli.NewClientLoginCommand())
	rootCmd.AddCommand(cli.NewClientAddBankCardCommand())
	rootCmd.AddCommand(cli.NewClientAddBinaryCommand())
	rootCmd.AddCommand(cli.NewClientAddTextCommand())
	rootCmd.AddCommand(cli.NewClientAddUserCommand())
	rootCmd.AddCommand(cli.NewClientListCommand())
	rootCmd.AddCommand(cli.NewClientSyncCommand())
	rootCmd.AddCommand(cli.NewClientInfoCommand(buildVersion, buildDate))

	return rootCmd.Execute()
}
