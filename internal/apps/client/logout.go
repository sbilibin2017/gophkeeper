package client

import (
	"fmt"
	"strings"

	clientHandlers "github.com/sbilibin2017/gophkeeper/internal/handlers/client"
	"github.com/spf13/cobra"
)

var (
	logoutAuthURL     string
	logoutTLSCertFile string
	logoutTLSKeyFile  string
	logoutToken       string
)

// injectable functions for testing/mocking
var (
	logoutHTTPFunc = clientHandlers.LogoutHTTP
	logoutGRPCFunc = clientHandlers.LogoutGRPC
)

func RegisterLogoutCommand(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "logout",
		Short:   "Logout the current user",
		Long:    "Logout the current user and invalidate the session token.",
		Example: `gophkeeper logout --auth-url https://example.com --token your-token --tls-client-cert cert.pem --tls-client-key key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case strings.HasPrefix(logoutAuthURL, "grpc://"):
				err := logoutGRPCFunc(cmd.Context(), logoutToken, logoutAuthURL, logoutTLSCertFile, logoutTLSKeyFile)
				if err != nil {
					return fmt.Errorf("logout failed: %w", err)
				}
			case strings.HasPrefix(logoutAuthURL, "http://"), strings.HasPrefix(logoutAuthURL, "https://"):
				err := logoutHTTPFunc(cmd.Context(), logoutToken, logoutAuthURL, logoutTLSCertFile, logoutTLSKeyFile)
				if err != nil {
					return fmt.Errorf("logout failed: %w", err)
				}
			default:
				return fmt.Errorf("unsupported auth URL scheme, must start with grpc://, http:// or https://")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&logoutAuthURL, "auth-url", "", "Authentication server URL")
	cmd.Flags().StringVar(&logoutTLSCertFile, "tls-client-cert", "", "Path to TLS client certificate file")
	cmd.Flags().StringVar(&logoutTLSKeyFile, "tls-client-key", "", "Path to TLS client key file")
	cmd.Flags().StringVar(&logoutToken, "token", "", "Session token to logout")

	cmd.MarkFlagRequired("auth-url")
	cmd.MarkFlagRequired("tls-client-cert")
	cmd.MarkFlagRequired("tls-client-key")
	cmd.MarkFlagRequired("token")

	root.AddCommand(cmd)
}
