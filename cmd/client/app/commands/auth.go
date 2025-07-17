package commands

import (
	"context"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
)

// RegisterRegisterCommand registers the "register" command to the root command.
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
  gophkeeper register --username bob --password 'S3cr3t!' --auth-url https://auth.example.com --tls-client-cert /path/to/cert.pem --tls-client-key /path/to/key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := client.ValidateRegisterUsername(username); err != nil {
				return err
			}
			if err := client.ValidateRegisterPassword(password); err != nil {
				return err
			}

			return runAuth(cmd, ctx, username, password, authURL, tlsClientCert, tlsClientKey)
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for authentication")
	cmd.Flags().StringVar(&password, "password", "", "Password for authentication")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file (optional)")

	root.AddCommand(cmd)
}

// RegisterLoginCommand registers the "login" command to the root cobra command.
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
  gophkeeper login --username bob --password 'S3cr3t!' --auth-url https://auth.example.com --tls-client-cert /path/to/cert.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := client.ValidateLoginUsername(username); err != nil {
				return err
			}
			if err := client.ValidateLoginPassword(password); err != nil {
				return err
			}

			return runAuth(cmd, ctx, username, password, authURL, tlsClientCert, "")
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for authentication")
	cmd.Flags().StringVar(&password, "password", "", "Password for authentication")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")

	root.AddCommand(cmd)
}

// runAuth handles shared logic for registration and login commands.
func runAuth(
	cmd *cobra.Command,
	ctx context.Context,
	username string,
	password string,
	authURL string,
	tlsClientCert string,
	tlsClientKey string,
) error {
	// 1. Setup client config (DB, HTTP/gRPC client)
	cfg, err := config.NewConfig(authURL, tlsClientCert, tlsClientKey)
	if err != nil {
		return fmt.Errorf("failed to create client config: %w", err)
	}

	// 2. Create required DB tables
	if err := client.CreateBinaryRequestTable(cfg.DB); err != nil {
		return fmt.Errorf("failed to create binary request table: %w", err)
	}
	if err := client.CreateTextRequestTable(cfg.DB); err != nil {
		return fmt.Errorf("failed to create text request table: %w", err)
	}
	if err := client.CreateUsernamePasswordRequestTable(cfg.DB); err != nil {
		return fmt.Errorf("failed to create username-password request table: %w", err)
	}
	if err := client.CreateBankCardRequestTable(cfg.DB); err != nil {
		return fmt.Errorf("failed to create bank card request table: %w", err)
	}

	// 3. Build auth request
	req := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	// 4. Authenticate via HTTP if available
	if cfg.HTTPClient != nil {
		resp, err := client.AuthHTTP(ctx, cfg.HTTPClient, req)
		if err != nil {
			return fmt.Errorf("http auth failed: %w", err)
		}
		cmd.Println(resp.Token)
		return nil
	}

	// 5. Authenticate via gRPC if available
	if cfg.GRPCClient != nil {
		grpcClient := pb.NewAuthServiceClient(cfg.GRPCClient)
		resp, err := client.AuthGRPC(ctx, grpcClient, req)
		if err != nil {
			return fmt.Errorf("grpc auth failed: %w", err)
		}
		cmd.Println(resp.Token)
		return nil
	}

	return fmt.Errorf("no HTTP or gRPC client available")
}
