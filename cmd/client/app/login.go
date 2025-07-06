package app

import (
	"github.com/spf13/cobra"
)

// newLoginCommand creates a cobra.Command for authenticating an existing user.
func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login --username <username> --password <password> [--server-url <url>]",
		Short: "Authenticate an existing user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	cmd.Flags().StringP("server-url", "s", "", "Server URL (optional)")
	cmd.Flags().StringP("username", "u", "", "Username (required)")
	cmd.Flags().StringP("password", "p", "", "User password (required)")

	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}
