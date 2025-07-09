package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newConfigureCommand creates the `configure` Cobra command,
// which allows the user to configure the client with a JWT token and server URL.
//
// Usage:
//
//	client configure --token <jwt_token> --server-url <url>
//
// The token and server URL are saved to environment variables GOPHKEEPER_TOKEN and GOPHKEEPER_SERVER_URL respectively.
// If the --token flag is not provided, the command returns an error.
// Setting server-url is optional.
func newConfigureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure client with a JWT token and server URL",
		Long:  "Configure client settings by setting the JWT token and optionally the server URL.",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := cmd.Flags().GetString("token")
			if err != nil {
				return fmt.Errorf("failed to get token flag")
			}

			if token == "" {
				return fmt.Errorf("token is required")
			}

			serverURL, err := cmd.Flags().GetString("server-url")
			if err != nil {
				return fmt.Errorf("failed to get server-url flag")
			}

			if err := os.Setenv("GOPHKEEPER_TOKEN", token); err != nil {
				return fmt.Errorf("failed to set GOPHKEEPER_TOKEN env: %w", err)
			}

			if serverURL != "" {
				if err := os.Setenv("GOPHKEEPER_SERVER_URL", serverURL); err != nil {
					return fmt.Errorf("failed to set GOPHKEEPER_SERVER_URL env: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().String("token", "", "JWT token")
	cmd.Flags().String("server-url", "", "Server URL")

	return cmd
}
