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
		Use:   "auth-logout",
		Short: "Logout the current user",
		Long:  `Logout current user by invalidating token.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

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
					return fmt.Errorf("gRPC logout failed: %w", err)
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
