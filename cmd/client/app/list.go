package app

import (
	"github.com/spf13/cobra"
)

// newListCommand creates a cobra.Command for listing stored data/secrets.
func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [--server-url <url>] [--secret-type <type>] [--sort <key>]",
		Short: "List stored secrets with optional filtering by type and sorting",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringP("server-url", "s", "", "Server URL (optional, fallback to config)")
	cmd.Flags().StringP("secret-type", "t", "", "Filter results by secret type")
	cmd.Flags().StringP("sort", "r", "", "Sort results by the specified key")

	return cmd
}
