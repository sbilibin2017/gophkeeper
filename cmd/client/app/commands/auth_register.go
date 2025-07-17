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
		Use:   "auth-register",
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
				return errors.New("unsupported scheme: " + schm)
			}

			cfg, err := configs.NewClientConfig(opts...)
			if err != nil {
				return err
			}

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

			req := &models.RegisterRequest{
				Username: username,
				Password: password,
			}

			if cfg.HTTPClient != nil {
				err := client.RegisterHTTP(ctx, cfg.HTTPClient, req)
				if err != nil {
					return err
				}
			}

			if cfg.GRPCClient != nil {
				registerClient := pb.NewRegisterServiceClient(cfg.GRPCClient)
				err := client.RegisterGRPC(ctx, registerClient, req)
				if err != nil {
					return err
				}
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
