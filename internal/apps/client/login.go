package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
	"github.com/spf13/cobra"
)

// Define injectable functions with defaults for easier testing/mocking
var (
	loginHTTPFunc = loginHTTP
	loginGRPCFunc = loginGRPC
)

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

func loginHTTP(
	ctx context.Context,
	username, password, authURL, tlsCertFile, tlsKeyFile string,
) (*models.AuthResponse, error) {
	config, err := configs.NewClientConfig(configs.WithAuthURL(authURL, tlsCertFile, tlsKeyFile))
	if err != nil {
		return nil, err
	}
	if config.HTTPClient == nil {
		return nil, fmt.Errorf("HTTP client is not configured for URL: %s", authURL)
	}

	authReq := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	facade := auth.NewLoginHTTPFacade(config.HTTPClient)

	resp, err := facade.Login(ctx, authReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func loginGRPC(
	ctx context.Context,
	username, password, authURL, tlsCertFile, tlsKeyFile string,
) (*models.AuthResponse, error) {
	config, err := configs.NewClientConfig(configs.WithAuthURL(authURL, tlsCertFile, tlsKeyFile))
	if err != nil {
		return nil, err
	}
	if config.GRPCClient == nil {
		return nil, fmt.Errorf("gRPC client is not configured for URL: %s", authURL)
	}

	authReq := &models.AuthRequest{
		Username: username,
		Password: password,
	}

	grpcClient := pb.NewAuthServiceClient(config.GRPCClient)
	facade := auth.NewLoginGRPCFacade(grpcClient) // Also create this facade similar to RegisterGRPCFacade

	resp, err := facade.Login(ctx, authReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
