package auth

import (
	"fmt"
	"strings"

	authHandlers "github.com/sbilibin2017/gophkeeper/internal/client/handlers/auth"
	"github.com/sbilibin2017/gophkeeper/internal/client/models"
	"github.com/spf13/cobra"
)

// Injectable functions for testing/mocking
var (
	loginHTTPFunc = authHandlers.LoginHTTP
	loginGRPCFunc = authHandlers.LoginGRPC
)

// RegisterLoginCommand adds the 'login' command to the provided root Cobra command.
//
// The login command authenticates a user by username and password against an authentication server.
// It supports both HTTP(S) and gRPC protocols, selectable based on the 'auth-url' flag prefix.
// The command requires TLS client certificate and key files for secure communication.
//
// Usage example:
//
//	gophkeeper login --username alice --password "S3cr3tPass!" --auth-url https://example.com --tls-client-cert cert.pem --tls-client-key key.pem
//
// Flags:
//
//	--username          Username for login (required)
//	--password          Password for login (required)
//	--auth-url          Authentication server URL (required, must start with grpc://, http://, or https://)
//	--tls-client-cert   Path to TLS client certificate file (required)
//	--tls-client-key    Path to TLS client key file (required)
func RegisterLoginCommand(root *cobra.Command) {
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
			var resp *models.AuthResponse
			var err error

			ctx := cmd.Context()

			switch {
			case strings.HasPrefix(authURL, "grpc://"):
				resp, err = loginGRPCFunc(ctx, username, password, authURL, tlsCertFile, tlsKeyFile)
			case strings.HasPrefix(authURL, "http://"), strings.HasPrefix(authURL, "https://"):
				resp, err = loginHTTPFunc(ctx, username, password, authURL, tlsCertFile, tlsKeyFile)
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
