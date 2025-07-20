package app

import (
	"fmt"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/client/handlers"
	"github.com/spf13/cobra"
)

// Injectable functions for testing/mocking
var (
	logoutHTTPFunc = handlers.LogoutHTTP
	logoutGRPCFunc = handlers.LogoutGRPC
)

// RegisterLogoutCommand adds the 'logout' command to the provided root Cobra command.
//
// The 'logout' command logs out the current user by invalidating the session token on the
// authentication server. It supports communication over HTTP(S) and gRPC based on the URL scheme.
//
// Flags:
//
//	--auth-url          Authentication server URL (required). Must start with grpc://, http://, or https://
//	--tls-client-cert   Path to TLS client certificate file (required)
//	--tls-client-key    Path to TLS client key file (required)
//	--token             Session token to logout (required)
//
// Example usage:
//
//	gophkeeper logout --auth-url https://example.com --token your-token --tls-client-cert cert.pem --tls-client-key key.pem
func RegisterLogoutCommand(root *cobra.Command) {
	var (
		authURL     string
		tlsCertFile string
		tlsKeyFile  string
		token       string
	)

	cmd := &cobra.Command{
		Use:     "logout",
		Short:   "Logout the current user",
		Long:    "Logout the current user and invalidate the session token.",
		Example: `gophkeeper logout --auth-url https://example.com --token your-token --tls-client-cert cert.pem --tls-client-key key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case strings.HasPrefix(authURL, "grpc://"):
				err := logoutGRPCFunc(cmd.Context(), token, authURL, tlsCertFile, tlsKeyFile)
				if err != nil {
					return fmt.Errorf("logout failed: %w", err)
				}
			case strings.HasPrefix(authURL, "http://"), strings.HasPrefix(authURL, "https://"):
				err := logoutHTTPFunc(cmd.Context(), token, authURL, tlsCertFile, tlsKeyFile)
				if err != nil {
					return fmt.Errorf("logout failed: %w", err)
				}
			default:
				return fmt.Errorf("unsupported auth URL scheme, must start with grpc://, http:// or https://")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication server URL")
	cmd.Flags().StringVar(&tlsCertFile, "tls-client-cert", "", "Path to TLS client certificate file")
	cmd.Flags().StringVar(&tlsKeyFile, "tls-client-key", "", "Path to TLS client key file")
	cmd.Flags().StringVar(&token, "token", "", "Session token to logout")

	cmd.MarkFlagRequired("auth-url")
	cmd.MarkFlagRequired("tls-client-cert")
	cmd.MarkFlagRequired("tls-client-key")
	cmd.MarkFlagRequired("token")

	root.AddCommand(cmd)
}
