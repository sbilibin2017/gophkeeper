package client

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewListCommand returns a cobra command stub that lists stored entries.
// Currently, the command is not implemented and prints a placeholder message.
func NewListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List stored entries",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("list command is not implemented yet")
			return nil
		},
	}
}
