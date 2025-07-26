package server

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewServerCommand returns a cobra command that starts the GophKeeper server.
// Currently, this is a placeholder implementation and does not run a real server.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Run the GophKeeper server",
		Long:  "Start the GophKeeper server to handle client requests and manage secure storage.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("server command is not implemented yet")
			return nil
		},
	}
}
