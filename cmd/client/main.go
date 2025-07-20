package main

import (
	"log"

	"github.com/sbilibin2017/gophkeeper/internal/client/app"
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

	app.RegisterRegisterCommand(rootCmd)

	return rootCmd.Execute()
}
