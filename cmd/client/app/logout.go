package app

import "github.com/spf13/cobra"

// newLogoutCommand creates a cobra.Command for logging out the current user.
func newLogoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
