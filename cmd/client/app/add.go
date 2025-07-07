package app

import (
	"github.com/spf13/cobra"
)

// newAddCommand creates a cobra command for adding new data/secrets
func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add --server-url <url> [--file <path>] [--interactive]",
		Short: "Add new data/secrets to the client from a file or interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringP("server-url", "s", "", "Server URL")
	cmd.Flags().StringP("file", "f", "", "Input file path")
	cmd.Flags().BoolP("interactive", "i", false, "Enable interactive input mode")

	return cmd
}
