package commands

import (
	"fmt"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
)

// RegisterLoginCommand registers the "login" command to the root cobra command.
// The "login" command authenticates a user by username and password using either HTTP or gRPC.
// It validates inputs, prepares the client configuration, ensures necessary DB tables exist,
// then sends the authentication request. On success, it prints the authentication token.
//
// Usage:
//
//	login --username USERNAME --password PASSWORD --auth-url URL [--tls-client-cert PATH]
//
// Example:
//
//	gophkeeper login --username alice --password 'P@ssw0rd123' --auth-url https://auth.example.com
func RegisterLoginCommand(root *cobra.Command) {
	var (
		username      string
		password      string
		authURL       string
		tlsClientCert string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate user",
		Long: `Authenticate a user by providing a username and password.

This command validates the credentials, initializes client configuration,
creates necessary database tables, and sends an authentication request to
the specified auth service endpoint using HTTP or gRPC protocols.

Upon successful authentication, an authentication token will be printed.`,
		Example: `  
  # Authenticate user with TLS client certificate for secure communication
  gophkeeper login --username bob --password 'S3cr3t!' --auth-url https://auth.example.com --tls-client-cert /path/to/cert.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Validate username and password for login
			if err := client.ValidateLoginUsername(username); err != nil {
				return fmt.Errorf("invalid username: %w", err)
			}
			if err := client.ValidateLoginPassword(password); err != nil {
				return fmt.Errorf("invalid password: %w", err)
			}

			// Create client config (including DB, HTTP/gRPC clients)
			config, err := config.NewConfig(authURL, tlsClientCert, "")
			if err != nil {
				return err
			}

			// Create required DB tables
			if err := client.CreateBinaryRequestTable(config.DB); err != nil {
				return fmt.Errorf("failed to create binary request table: %w", err)
			}
			if err := client.CreateTextRequestTable(config.DB); err != nil {
				return fmt.Errorf("failed to create text request table: %w", err)
			}
			if err := client.CreateUsernamePasswordRequestTable(config.DB); err != nil {
				return fmt.Errorf("failed to create username-password request table: %w", err)
			}
			if err := client.CreateBankCardRequestTable(config.DB); err != nil {
				return fmt.Errorf("failed to create bank card request table: %w", err)
			}

			req := &models.AuthRequest{
				Username: username,
				Password: password,
			}

			// Use HTTP client if configured
			if config.HTTPClient != nil {
				resp, err := client.AuthHTTP(ctx, config.HTTPClient, req)
				if err != nil {
					return fmt.Errorf("http login failed: %w", err)
				}
				cmd.Println(resp.Token)
				return nil
			}

			// Use gRPC client if configured
			if config.GRPCClient != nil {
				grpcClient := pb.NewAuthServiceClient(config.GRPCClient)
				resp, err := client.AuthGRPC(ctx, grpcClient, req)
				if err != nil {
					return fmt.Errorf("grpc login failed: %w", err)
				}
				cmd.Println(resp.Token)
				return nil
			}

			return fmt.Errorf("no HTTP or gRPC client available")
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for authentication")
	cmd.Flags().StringVar(&password, "password", "", "Password for authentication")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")

	root.AddCommand(cmd)
}
