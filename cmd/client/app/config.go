package app

import (
	"github.com/spf13/cobra"
)

// newClientConfigGetCommand creates a command to show the current client configuration.
func newClientConfigGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Display the current client configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

// newClientConfigSetCommand creates a command to set a key-value pair in the client configuration.
func newClientConfigSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a client configuration key to a specified value",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return cobra.MinimumNArgs(2)(cmd, args)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

// newClientConfigUnsetCommand creates a command to remove a key from the client configuration.
func newClientConfigUnsetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Remove a key from the client configuration",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cobra.MinimumNArgs(1)(cmd, args)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}
