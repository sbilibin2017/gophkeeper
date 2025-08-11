package list

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/spf13/cobra"
)

// NewCommand returns the "list" CLI command.
func NewCommand() *cobra.Command {
	var (
		serverURL string
		token     string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all secrets",
		Long: `Fetch and decrypt all secrets from the Gophkeeper backend.
Requires an authentication token.

Examples of usage and supported server URL schemes:
- http:// for HTTP
- grpc:// for gRPC
`,
		Example: `  # List secrets using HTTP
  gophkeeper list --token <token>

  # List secrets using gRPC
  gophkeeper list --token <token> --server-url grpc://localhost:50051
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			addr := address.New(serverURL)
			if addr.Address == "" || addr.Scheme == "" {
				return errors.New("invalid server URL format")
			}

			var output string
			var err error

			switch addr.Scheme {
			case address.SchemeHTTP, address.SchemeHTTPS:
				output, err = RunHTTP(ctx, token, addr.Address)
			case address.SchemeGRPC:
				output, err = RunGRPC(ctx, token, addr.Address)
			default:
				return address.ErrUnsupportedScheme
			}

			if err != nil {
				return fmt.Errorf("failed to list secrets: %w", err)
			}

			cmd.Println(output)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "Server address (scheme://host:port)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	return cmd
}
