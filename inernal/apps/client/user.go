// Package client contains CLI commands for the GophKeeper client application.
package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewRegisterCommand returns a cobra command stub that registers a new user.
// Currently, the command is not implemented and prints a placeholder message.
func NewRegisterCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("register command is not implemented yet")
			return nil
		},
	}
}

// NewLoginCommand returns a cobra command stub that logs in an existing user.
// Currently, the command is not implemented and prints a placeholder message.
func NewLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login user",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("login command is not implemented yet")
			return nil
		},
	}
}
