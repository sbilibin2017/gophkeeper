package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// newConfigureCommand creates the `configure` Cobra command,
// which allows the user to configure the client with a JWT token.
//
// Usage:
//
//	client configure --token <jwt_token>
//
// The token is saved to the file located at $HOME/.gophkeeper.
// If the --token flag is not provided, the command returns an error.
// The token file is written with permission 0600 (user read/write only).
func newConfigureCommand() *cobra.Command {
	var token string

	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure client with a JWT token",
		Long:  "Configure client settings by setting the JWT token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" {
				return fmt.Errorf("token is required")
			}

			configPath := os.ExpandEnv("$HOME/.gophkeeper")
			err := os.WriteFile(configPath, []byte(token), 0600)
			if err != nil {
				return fmt.Errorf("failed to save token")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "JWT token to configure the client")

	return cmd
}
