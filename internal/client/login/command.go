package login

import (
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

// NewCommand returns the "login" CLI command.
func NewCommand() *cobra.Command {
	var (
		serverURL string
		username  string
		password  string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in a user",
		Long: `Authenticate a user with the Gophkeeper service.
This command sends the username and password to the backend server
using either HTTP or gRPC protocols, depending on the specified server URL scheme.`,
		Example: `  # Login with default HTTP server
  gophkeeper login --username alice --password secret  

  # Login using gRPC server
  gophkeeper login --username charlie --password secret --server-url grpc://localhost:50051
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
	cmd.Flags().StringVar(&username, "username", "", "Username for login")
	cmd.Flags().StringVar(&password, "password", "", "Password for login")

	return cmd
}
