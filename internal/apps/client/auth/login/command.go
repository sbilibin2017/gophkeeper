package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

// RegisterCommand adds the "login" subcommand to the provided Cobra root command.
// This command allows a user to authenticate against an HTTP or gRPC service.
//
// Parameters:
//   - root: The root Cobra command to which the "login" subcommand will be added.
//   - runHTTPFunc: A function that performs HTTP-based authentication using context, username, and password.
//   - runGRPCFunc: A function that performs gRPC-based authentication using context, username, and password.
//
// The subcommand accepts the following flags:
//
//	--username         Required: The username to login with.
//	--password         Required: The password for the specified user.
//	--auth-url         Required: The URL of the authentication service (must begin with grpc://, http://, or https://).
//	--tls-client-cert  Required: Path to the client TLS certificate file.
//	--tls-client-key   Required: Path to the client TLS key file.
//
// Example usage:
//
//	gophkeeper login \
//	  --username alice \
//	  --password S3cr3tPass! \
//	  --auth-url https://example.com \
//	  --tls-client-cert cert.pem \
//	  --tls-client-key key.pem
func RegisterCommand(
	root *cobra.Command,
	runHTTPFunc func(ctx context.Context, username, password string) (*models.AuthResponse, error),
	runGRPCFunc func(ctx context.Context, username, password string) (*models.AuthResponse, error),
) {
	var (
		username    string
		password    string
		authURL     string
		tlsCertFile string
		tlsKeyFile  string
	)

	cmd := &cobra.Command{
		Use:     "login",
		Short:   "Login a user",
		Long:    "Authenticate a user by username, password, and authentication details to obtain a session token.",
		Example: `gophkeeper login --username alice --password "S3cr3tPass!" --auth-url https://example.com --tls-client-cert cert.pem --tls-client-key key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			var resp *models.AuthResponse
			var err error

			switch {
			case strings.HasPrefix(authURL, "grpc://"):
				resp, err = runGRPCFunc(ctx, username, password)
			case strings.HasPrefix(authURL, "http://"), strings.HasPrefix(authURL, "https://"):
				resp, err = runHTTPFunc(ctx, username, password)
			default:
				return fmt.Errorf("unsupported auth URL scheme, must start with grpc://, http:// or https://")
			}

			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			cmd.Println(resp.Token)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for login")
	cmd.Flags().StringVar(&password, "password", "", "Password for login")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication server URL")
	cmd.Flags().StringVar(&tlsCertFile, "tls-client-cert", "", "Path to TLS client certificate file")
	cmd.Flags().StringVar(&tlsKeyFile, "tls-client-key", "", "Path to TLS client key file")

	cmd.MarkFlagRequired("username")
	cmd.MarkFlagRequired("password")
	cmd.MarkFlagRequired("auth-url")
	cmd.MarkFlagRequired("tls-client-cert")
	cmd.MarkFlagRequired("tls-client-key")

	root.AddCommand(cmd)
}
