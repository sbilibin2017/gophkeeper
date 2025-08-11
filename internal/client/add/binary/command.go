package binary

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/spf13/cobra"
)

// NewCommand returns the "binary" CLI command.
func NewCommand() *cobra.Command {
	var (
		serverURL  string
		token      string
		secretName string
		dataPath   string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "binary",
		Short: "Save a new binary secret",
		Long: `Encrypt and save binary secret data to the Gophkeeper backend.
Requires an authentication token.

Examples of usage and supported server URL schemes:
- http:// for HTTP
- grpc:// for gRPC
`,
		Example: `  # Save a binary secret using HTTP
  gophkeeper binary --token <token> --secret-name mybinary --data-path /path/to/file.bin

  # Save a binary secret using gRPC
  gophkeeper binary --token <token> --server-url grpc://localhost:50051 --secret-name mybinary --data-path /path/to/file.bin
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
				err = RunHTTP(ctx, token, addr.Address, secretName, dataPath, meta)
			case address.SchemeGRPC:
				err = RunGRPC(ctx, token, addr.Address, secretName, dataPath, meta)
			default:
				return address.ErrUnsupportedScheme
			}

			if err != nil {
				return fmt.Errorf("failed to save binary secret: %w", err)
			}

			cmd.Println("Binary secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "Server address (scheme://host:port)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "Path to the binary file to save")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}
