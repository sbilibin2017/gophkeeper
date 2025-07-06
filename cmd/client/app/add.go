package app

import (
	"github.com/spf13/cobra"
)

// newAddCommand creates a cobra.Command for adding new data/secrets to the client
// from a file or interactively.
func newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [--file <path>] [--interactive]",
		Short: "Add new data/secrets to the client from a file or interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Path to input file")
	cmd.Flags().BoolP("interactive", "i", false, "Enable interactive input mode")

	return cmd
}
