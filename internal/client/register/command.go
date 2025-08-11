package register

import (
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

// NewCommand returns the "register" CLI command.
func NewCommand() *cobra.Command {
	var (
		serverURL string
		username  string
		password  string
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Long: `Register a new user with the Gophkeeper service. 
This command sends the username and password to the backend server
using either HTTP or gRPC protocols, depending on the specified server URL scheme.

Examples of usage and supported server URL schemes:
- http:// for HTTP
- grpc:// for gRPC
`,
		Example: `  # Register with default HTTP server
  gophkeeper register --username alice --password secret  

  # Register using gRPC server
  gophkeeper register --username charlie --password secret --server-url grpc://localhost:50051
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			req := &models.AuthRequest{
				Username: username,
				Password: password,
			}

			addr := address.New(serverURL)

			if addr.Address == "" || addr.Scheme == "" {
				return errors.New("invalid server URL format")
			}

			var resp *models.AuthResponse
			var err error

			switch addr.Scheme {
			case address.SchemeHTTP, address.SchemeHTTPS:
				resp, err = RunHTTP(ctx, addr.Address, req)
			case address.SchemeGRPC:
				resp, err = RunGRPC(ctx, addr.Address, req)
			default:
				return address.ErrUnsupportedScheme
			}

			if err != nil {
				return err
			}

			cmd.Println(resp.Token)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "Server address (scheme://host:port)")
	cmd.Flags().StringVar(&username, "username", "", "Username for registration")
	cmd.Flags().StringVar(&password, "password", "", "Password for registration")

	return cmd
}
