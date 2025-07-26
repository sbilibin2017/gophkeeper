// Package client contains CLI commands for the GophKeeper client application.
package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewAddBankCardCommand returns a cobra command stub for adding a bank card entry.
// Currently, the command is not implemented and prints a placeholder message.
func NewAddBankCardCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-bankcard",
		Short: "Add a new bank card entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("add-bankcard command is not implemented yet")
			return nil
		},
	}
}

// NewAddBinaryCommand returns a cobra command stub for adding binary data.
// Currently, the command is not implemented and prints a placeholder message.
func NewAddBinaryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-binary",
		Short: "Add a new binary data entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("add-binary command is not implemented yet")
			return nil
		},
	}
}

// NewAddTextCommand returns a cobra command stub for adding a text entry.
// Currently, the command is not implemented and prints a placeholder message.
func NewAddTextCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-text",
		Short: "Add a new text entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("add-text command is not implemented yet")
			return nil
		},
	}
}

// NewAddUserCommand returns a cobra command stub for adding a user record.
// Currently, the command is not implemented and prints a placeholder message.
func NewAddUserCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-user",
		Short: "Add a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("add-user command is not implemented yet")
			return nil
		},
	}
}
