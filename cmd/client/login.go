package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/apps/client/login"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the service",
	Long:  `Perform user login via gRPC or HTTP depending on the auth-url protocol prefix.`,
	Example: `  
# Login using gRPC protocol
login --auth-url grpc://localhost:50051 --username user1 --password pass123 --tls-cert path/to/cert.pem --tls-key path/to/key.pem

# Login using HTTP protocol
login --auth-url http://localhost:8080 --username user1 --password pass123 --tls-cert path/to/cert.pem --tls-key path/to/key.pem
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
			resp, err := login.RunLoginGRPC(ctx, authURL, tlsCertFile, tlsKeyFile, username, password)
			if err != nil {
				return fmt.Errorf("gRPC login failed: %w", err)
			}
			cmd.Print(resp.Token)
		case scheme.HTTP, scheme.HTTPS:
			resp, err := login.RunLoginHTTP(ctx, authURL, tlsCertFile, tlsKeyFile, username, password)
			if err != nil {
				return fmt.Errorf("HTTP login failed: %w", err)
			}
			cmd.Print(resp.Token)
		default:
			return fmt.Errorf("unsupported protocol scheme: %s", protocol)
		}

		return nil
	},
}

func init() {
	loginCmd.Flags().String("username", "", "Username for login")
	loginCmd.Flags().String("password", "", "Password for login")

	loginCmd.Flags().String("auth-url", "", "Auth server URL with protocol prefix, e.g. grpc://host:port or http://host:port")
	loginCmd.Flags().String("tls-cert", "", "Path to TLS certificate file")
	loginCmd.Flags().String("tls-key", "", "Path to TLS key file")
}
