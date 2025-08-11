package text

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/spf13/cobra"
)

// NewCommand returns the "text" CLI command.
func NewCommand() *cobra.Command {
	var (
		serverURL  string
		token      string
		secretName string
		data       string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "text",
		Short: "Save a new text secret",
		Long: `Encrypt and save text secret data to the Gophkeeper backend.
Requires an authentication token.

Examples of usage and supported server URL schemes:
- http:// for HTTP
- grpc:// for gRPC
`,
		Example: `  # Save a text secret using HTTP
  gophkeeper text --token <token> --secret-name mytext --data "This is a secret note"

  # Save a text secret using gRPC
  gophkeeper text --token <token> --server-url grpc://localhost:50051 --secret-name mytext --data "This is a secret note"
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
				err = RunHTTP(ctx, token, addr.Address, secretName, data, meta)
			case address.SchemeGRPC:
				err = RunGRPC(ctx, token, addr.Address, secretName, data, meta)
			default:
				return address.ErrUnsupportedScheme
			}

			if err != nil {
				return fmt.Errorf("failed to save text secret: %w", err)
			}

			cmd.Println("Text secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "Server address (scheme://host:port)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&data, "data", "", "Text secret data")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}
