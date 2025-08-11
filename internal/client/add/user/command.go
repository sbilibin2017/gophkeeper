package user

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/spf13/cobra"
)

// NewCommand returns the "usersecret" CLI command.
func NewCommand() *cobra.Command {
	var (
		serverURL  string
		token      string
		secretName string
		username   string
		password   string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "usersecret",
		Short: "Save a new user secret",
		Long: `Encrypt and save user credentials to the Gophkeeper backend.
Requires an authentication token.

Examples of usage and supported server URL schemes:
- http:// for HTTP
- grpc:// for gRPC
`,
		Example: `  # Save a user secret using HTTP
  gophkeeper usersecret --token <token> --secret-name myuser --username alice --password "pass123"

  # Save a user secret using gRPC
  gophkeeper usersecret --token <token> --server-url grpc://localhost:50051 --secret-name myuser --username alice --password "pass123"
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			addr := address.New(serverURL)
			if addr.Address == "" || addr.Scheme == "" {
				return errors.New("invalid server URL format")
			}

			var err error
			switch addr.Scheme {
			case address.SchemeHTTP, address.SchemeHTTPS:
				err = RunHTTP(ctx, token, addr.Address, secretName, username, password, meta)
			case address.SchemeGRPC:
				err = RunGRPC(ctx, token, addr.Address, secretName, username, password, meta)
			default:
				return address.ErrUnsupportedScheme
			}

			if err != nil {
				return fmt.Errorf("failed to save user secret: %w", err)
			}

			cmd.Println("User secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "Server address (scheme://host:port)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&username, "username", "", "Username for the secret")
	cmd.Flags().StringVar(&password, "password", "", "Password for the secret")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}
