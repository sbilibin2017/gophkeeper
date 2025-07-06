package app

import (
	"github.com/spf13/cobra"
)

// newGetCommand creates a cobra.Command for retrieving a specific secret/data by ID from the server.
func newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [--secret-id <id>] [--server-url <url>]",
		Short: "Retrieve specific data/secret from the server by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringP("secret-id", "i", "", "Identifier of the data to retrieve (optional)")
	cmd.Flags().StringP("server-url", "s", "", "Server URL (optional)")

	return cmd
}
