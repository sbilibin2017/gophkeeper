package commands

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config" // import your config package here
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/validation"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
)

// RegisterLoginCommand adds the "login" CLI command to the root command.
// It authenticates a user by username and password with HTTP-first then gRPC fallback logic.
// The command also creates necessary DB tables before login.
func RegisterLoginCommand(root *cobra.Command) {
	var (
		username      string
		password      string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
	)

	cmd := &cobra.Command{
		Use:   "auth-login",
		Short: "Authenticate user",
		Long:  `Authenticate user by username and password.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := validation.ValidateString("username", username); err != nil {
				return err
			}
			if err := validation.ValidateString("password", password); err != nil {
				return err
			}

			// Use centralized client config creation
			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			// Create necessary DB tables
			if err := client.CreateBinaryRequestTable(ctx, cfg.DB); err != nil {
				return fmt.Errorf("failed to create binary request table: %w", err)
			}
			if err := client.CreateTextRequestTable(ctx, cfg.DB); err != nil {
				return fmt.Errorf("failed to create text request table: %w", err)
			}
			if err := client.CreateUsernamePasswordRequestTable(ctx, cfg.DB); err != nil {
				return fmt.Errorf("failed to create username-password request table: %w", err)
			}
			if err := client.CreateBankCardRequestTable(ctx, cfg.DB); err != nil {
				return fmt.Errorf("failed to create bank card request table: %w", err)
			}

			req := &models.LoginRequest{
				Username: username,
				Password: password,
			}

			// Try HTTP login first if HTTP client available
			if cfg.HTTPClient != nil {
				resp, err := client.LoginHTTP(ctx, cfg.HTTPClient, req)
				if err == nil {
					cmd.Println(resp.Token)
					return nil
				}
				cmd.Printf("HTTP login failed: %v, trying gRPC fallback...\n", err)
			}

			// Try gRPC login if gRPC client available
			if cfg.GRPCClient != nil {
				loginClient := pb.NewLoginServiceClient(cfg.GRPCClient)
				resp, err := client.LoginGRPC(ctx, loginClient, req)
				if err != nil {
					return fmt.Errorf("gRPC login failed: %w", err)
				}
				cmd.Println(resp.Token)
				return nil
			}

			return errors.New("no HTTP or gRPC client available for login command")
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for login")
	cmd.Flags().StringVar(&password, "password", "", "Password for login")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file")

	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("auth-url")
	_ = cmd.MarkFlagRequired("tls-client-cert")
	_ = cmd.MarkFlagRequired("tls-client-key")

	root.AddCommand(cmd)
}
