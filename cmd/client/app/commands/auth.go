package commands

import (
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
)

// RegisterRegisterCommand adds the "register" CLI command to the root command.
// It registers a new user by username and password with HTTP-first then gRPC fallback logic.
// The command also creates necessary DB tables before registration.
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
		Long:  `Register a new user account by providing a username and password.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := client.ValidateRegisterUsername(username); err != nil {
				return err
			}
			if err := client.ValidateRegisterPassword(password); err != nil {
				return err
			}

			cfg, err := newAuthConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

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

			req := &models.RegisterRequest{
				Username: username,
				Password: password,
			}

			if cfg.HTTPClient != nil {
				resp, err := client.RegisterHTTP(ctx, cfg.HTTPClient, req)
				if err == nil {
					cmd.Println(resp.Token)
					return nil
				}
				cmd.Printf("HTTP register failed: %v, trying gRPC fallback...\n", err)
			}

			if cfg.GRPCClient != nil {
				registerClient := pb.NewRegisterServiceClient(cfg.GRPCClient)
				resp, err := client.RegisterGRPC(ctx, registerClient, req)
				if err != nil {
					return fmt.Errorf("grpc register failed: %w", err)
				}
				cmd.Println(resp.Token)
				return nil
			}

			return fmt.Errorf("no HTTP or gRPC client available for register command")
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Username for registration")
	cmd.Flags().StringVar(&password, "password", "", "Password for registration")
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
		Use:   "login",
		Short: "Authenticate user",
		Long:  `Authenticate user by username and password.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := client.ValidateLoginUsername(username); err != nil {
				return err
			}
			if err := client.ValidateLoginPassword(password); err != nil {
				return err
			}

			cfg, err := newAuthConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

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

			req := &models.LoginRequest{
				Username: username,
				Password: password,
			}

			if cfg.HTTPClient != nil {
				resp, err := client.LoginHTTP(ctx, cfg.HTTPClient, req)
				if err == nil {
					cmd.Println(resp.Token)
					return nil
				}
				cmd.Printf("HTTP login failed: %v, trying gRPC fallback...\n", err)
			}

			if cfg.GRPCClient != nil {
				loginClient := pb.NewLoginServiceClient(cfg.GRPCClient)
				resp, err := client.LoginGRPC(ctx, loginClient, req)
				if err != nil {
					return fmt.Errorf("grpc login failed: %w", err)
				}
				cmd.Println(resp.Token)
				return nil
			}

			return fmt.Errorf("no HTTP or gRPC client available for login command")
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

// RegisterLogoutCommand adds the "logout" CLI command to the root command.
// It logs out the current user by invalidating the authentication token,
// with HTTP-first then gRPC fallback logic.
func RegisterLogoutCommand(root *cobra.Command) {
	var (
		token         string
		authURL       string
		tlsClientCert string
		tlsClientKey  string
	)

	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout the current user",
		Long:  `Logout current user by invalidating token.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := newAuthConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			req := &models.LogoutRequest{
				Token: token,
			}

			if cfg.HTTPClient != nil {
				if err := client.LogoutHTTP(ctx, cfg.HTTPClient, req); err == nil {
					cmd.Println("Logout successful")
					return nil
				} else {
					cmd.Printf("HTTP logout failed: %v, trying gRPC fallback...\n", err)
				}
			}

			if cfg.GRPCClient != nil {
				logoutClient := pb.NewLogoutServiceClient(cfg.GRPCClient)
				if err := client.LogoutGRPC(ctx, logoutClient, req); err != nil {
					return fmt.Errorf("grpc logout failed: %w", err)
				}
				cmd.Println("Logout successful")
				return nil
			}

			return fmt.Errorf("no HTTP or gRPC client available for logout command")
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file")

	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("auth-url")
	_ = cmd.MarkFlagRequired("tls-client-cert")
	_ = cmd.MarkFlagRequired("tls-client-key")

	root.AddCommand(cmd)
}

// newConfig creates a new ClientConfig based on the given authentication URL
// and optional TLS client certificate and key files.
//
// The function determines the scheme from the authURL (HTTP, HTTPS, or gRPC)
// and configures the client accordingly:
// - For HTTP/HTTPS, it configures an HTTP client with optional TLS certificates.
// - For gRPC, it configures a gRPC client with optional TLS certificates.
//
// Returns an error if the URL scheme is unsupported or configuration fails.
func newAuthConfig(
	authURL string,
	tlsClientCert string,
	tlsClientKey string,
) (*configs.ClientConfig, error) {
	var opts []configs.ClientConfigOpt

	opts = append(opts, configs.WithClientConfigDB())

	schm := scheme.GetSchemeFromURL(authURL)

	switch schm {
	case scheme.HTTP, scheme.HTTPS:
		httpOpts := []clients.HTTPClientOption{}
		if tlsClientCert != "" && tlsClientKey != "" {
			httpOpts = append(httpOpts, clients.WithHTTPTLSClientCert(tlsClientCert, tlsClientKey))
		}
		opts = append(opts, configs.WithClientConfigHTTPClient(authURL, httpOpts...))

	case scheme.GRPC:
		grpcOpts := []clients.GRPCClientOption{}
		if tlsClientCert != "" && tlsClientKey != "" {
			grpcOpts = append(grpcOpts, clients.WithGRPCTLSClientCert(tlsClientCert, tlsClientKey))
		}
		opts = append(opts, configs.WithClientConfigGRPCClient(authURL, grpcOpts...))

	default:
		return nil, errors.New("unsupported scheme: " + schm)
	}

	return configs.NewClientConfig(opts...)
}
