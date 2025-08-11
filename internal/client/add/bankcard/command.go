package bankcard

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/spf13/cobra"
)

// NewCommand returns the "bankcard" CLI command.
func NewCommand() *cobra.Command {
	var (
		serverURL  string
		token      string
		secretName string
		number     string
		owner      string
		exp        string
		cvv        string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "bankcard",
		Short: "Save a new bank card secret",
		Long: `Encrypt and save bank card data to the Gophkeeper backend.
Requires an authentication token.

Examples of usage and supported server URL schemes:
- http:// for HTTP
- grpc:// for gRPC
`,
		Example: `  # Save a bank card using HTTP
  gophkeeper bankcard --token <token> --secret-name mycard --number 4111111111111111 --owner "Alice" --exp 12/25 --cvv 123

  # Save a bank card using gRPC
  gophkeeper bankcard --token <token> --server-url grpc://localhost:50051 --secret-name mycard --number 4111111111111111 --owner "Alice" --exp 12/25 --cvv 123
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
				err = RunHTTP(ctx, token, addr.Address, secretName, number, owner, exp, cvv, meta)
			case address.SchemeGRPC:
				err = RunGRPC(ctx, token, addr.Address, secretName, number, owner, exp, cvv, meta)
			default:
				return address.ErrUnsupportedScheme
			}

			if err != nil {
				return fmt.Errorf("failed to save bank card: %w", err)
			}

			cmd.Println("Bank card saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "Server address (scheme://host:port)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret")
	cmd.Flags().StringVar(&number, "number", "", "Bank card number")
	cmd.Flags().StringVar(&owner, "owner", "", "Bank card owner")
	cmd.Flags().StringVar(&exp, "exp", "", "Expiration date (MM/YY)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "CVV code")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	return cmd
}
