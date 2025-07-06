package app

import (
	"github.com/spf13/cobra"
)

// newSyncCommand creates a cobra.Command for synchronizing the client with the server.
// It supports optional flags for automatic conflict resolution and interactive resolution.
// The server URL can also be specified, falling back to the configuration if not provided.
func newSyncCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [--auto-resolve=client|server] [--interactive] [--server-url <url>]",
		Short: "Synchronize client with server and resolve conflicts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringP("auto-resolve", "a", "", "Automatically resolve conflicts using either 'client' or 'server'")
	cmd.Flags().BoolP("interactive", "i", false, "Enable interactive conflict resolution")
	cmd.Flags().StringP("server-url", "s", "", "Server URL (optional, fallback to config)")

	return cmd
}
