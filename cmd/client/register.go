package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/apps/client/register"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Long:  `Perform user registration via gRPC or HTTP depending on the auth-url protocol prefix.`,
	Example: `
# Register using gRPC protocol
register --auth-url grpc://localhost:50051 --username user1 --password pass123 --tls-cert path/to/cert.pem --tls-key path/to/key.pem

# Register using HTTP protocol
register --auth-url http://localhost:8080 --username user1 --password pass123 --tls-cert path/to/cert.pem --tls-key path/to/key.pem
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		authURL, err := cmd.Flags().GetString("auth-url")
		if err != nil {
			return err
		}
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			return err
		}
		password, err := cmd.Flags().GetString("password")
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
			resp, err := register.RunRegisterGRPC(ctx, authURL, tlsCertFile, tlsKeyFile, username, password)
			if err != nil {
				return fmt.Errorf("gRPC registration failed: %w", err)
			}
			cmd.Print(resp.Token)
		case scheme.HTTP, scheme.HTTPS:
			resp, err := register.RunRegisterHTTP(ctx, authURL, tlsCertFile, tlsKeyFile, username, password)
			if err != nil {
				return fmt.Errorf("HTTP registration failed: %w", err)
			}
			cmd.Print(resp.Token)
		default:
			return fmt.Errorf("unsupported protocol scheme: %s", protocol)
		}

		return nil
	},
}

func init() {
	registerCmd.Flags().String("username", "", "Username for registration")
	registerCmd.Flags().String("password", "", "Password for registration")

	registerCmd.Flags().String("auth-url", "", "Auth server URL with protocol prefix, e.g. grpc://host:port or http://host:port")
	registerCmd.Flags().String("tls-cert", "", "Path to TLS certificate file")
	registerCmd.Flags().String("tls-key", "", "Path to TLS key file")
}
