package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/apps/client/logout"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from the service",
	Long:  `Perform user logout via gRPC or HTTP depending on the auth-url protocol prefix.`,
	Example: `
# Logout using gRPC protocol
logout --auth-url grpc://localhost:50051 --token your_token_here --tls-cert path/to/cert.pem --tls-key path/to/key.pem

# Logout using HTTP protocol
logout --auth-url http://localhost:8080 --token your_token_here --tls-cert path/to/cert.pem --tls-key path/to/key.pem
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		authURL, err := cmd.Flags().GetString("auth-url")
		if err != nil {
			return err
		}
		token, err := cmd.Flags().GetString("token")
		if err != nil {
			return err
		}
		tlsCertFile, err := cmd.Flags().GetString("tls-cert")
		if err != nil {
			return err
		}
		tlsKeyFile, err := cmd.Flags().GetString("tls-key")
		if err != nil {
			return err
		}

		protocol := scheme.GetSchemeFromURL(authURL)
		if protocol == "" {
			return fmt.Errorf("unsupported or missing protocol scheme in auth-url")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		switch protocol {
		case scheme.GRPC:
			err := logout.RunLogoutGRPC(ctx, authURL, tlsCertFile, tlsKeyFile, token)
			if err != nil {
				return fmt.Errorf("gRPC logout failed: %w", err)
			}
		case scheme.HTTP, scheme.HTTPS:
			err := logout.RunLogoutHTTP(ctx, authURL, tlsCertFile, tlsKeyFile, token)
			if err != nil {
				return fmt.Errorf("HTTP logout failed: %w", err)
			}
		default:
			return fmt.Errorf("unsupported protocol scheme: %s", protocol)
		}

		return nil
	},
}

func init() {
	logoutCmd.Flags().String("token", "", "Authentication token for logout")

	logoutCmd.Flags().String("auth-url", "", "Auth server URL with protocol prefix, e.g. grpc://host:port or http://host:port")
	logoutCmd.Flags().String("tls-cert", "", "Path to TLS certificate file")
	logoutCmd.Flags().String("tls-key", "", "Path to TLS key file")
}
