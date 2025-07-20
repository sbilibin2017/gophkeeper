package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/internal/client/app"
	"github.com/sbilibin2017/gophkeeper/internal/client/app/auth"
	"github.com/sbilibin2017/gophkeeper/internal/client/app/bankcard"
)

// main is the entry point of the application.
// It runs the CLI command and logs any fatal errors.
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run initializes the root CLI command, registers subcommands,
// and executes the command tree.
//
// Returns an error if command execution fails.
func run() error {
	rootCmd := app.NewRootCommand()
	auth.RegisterRegisterCommand(rootCmd)
	bankcard.RegisterAddCommand(rootCmd)
	return rootCmd.Execute()
}
