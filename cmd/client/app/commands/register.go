package commands

import (
	"fmt"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
)

// RegisterRegisterCommand registers the "register" command to the root command.
// This command allows a user to register a new account by providing a username and password.
//
// Usage:
//
//	register --username USERNAME --password PASSWORD --auth-url URL [--tls-client-cert PATH] [--tls-client-key PATH]
//
// Example:
//
//	gophkeeper register --username alice --password 'P@ssw0rd123' --auth-url https://auth.example.com
func RegisterRegisterCommand(root *cobra.Command) {
	var (
		username      string
		password      string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		Long: `Register a new user account by providing a username and password.

This command validates the provided credentials according to registration rules,
sets up the client configuration including HTTP or gRPC clients,
creates necessary database tables, and sends a registration request to the auth service.

Upon successful registration, an authentication token will be printed.`,
		Example: `
  # Register a new user with TLS client cert and key for secure communication
  gophkeeper register --username bob --password 'S3cr3t!' --auth-url https://auth.example.com --tls-client-cert /path/to/cert.pem --tls-client-key /path/to/key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Validate username and password according to registration rules.
			if err := client.ValidateRegisterUsername(username); err != nil {
				return fmt.Errorf("invalid username: %w", err)
			}
			if err := client.ValidateRegisterPassword(password); err != nil {
				return fmt.Errorf("invalid password: %w", err)
			}

			// Create client configuration including HTTP or gRPC clients.
			config, err := config.NewConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			// Create necessary DB tables after client initialization.
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

			// Use HTTP client if available
			if config.HTTPClient != nil {
				resp, err := client.AuthHTTP(ctx, config.HTTPClient, req)
				if err != nil {
					return fmt.Errorf("http register failed: %w", err)
				}
				cmd.Println(resp.Token)
				return nil
			}

			// Use gRPC client if available
			if config.GRPCClient != nil {
				grpcClient := pb.NewAuthServiceClient(config.GRPCClient)
				resp, err := client.AuthGRPC(ctx, grpcClient, req)
				if err != nil {
					return fmt.Errorf("grpc register failed: %w", err)
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
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file (optional)")

	root.AddCommand(cmd)
}
