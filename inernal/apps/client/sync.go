package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSyncCommand returns a cobra command stub that synchronizes local data with the server.
// Currently, the command is not implemented and prints a placeholder message.
func NewSyncCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync local data with the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("sync command is not implemented yet")
			return nil
		},
	}
}
