package commands

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

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
		Long:  "Logout current user by invalidating token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			req := &models.LogoutRequest{
				Token: token,
			}

			if cfg.HTTPClient != nil {
				if err := client.LogoutHTTP(ctx, cfg.HTTPClient, req); err == nil {
					cmd.Println("Logout successful")
					return nil
				} else {
					cmd.Println("HTTP logout failed, trying gRPC fallback...")
				}
			}

			if cfg.GRPCClient != nil {
				logoutClient := pb.NewLogoutServiceClient(cfg.GRPCClient)
				if err := client.LogoutGRPC(ctx, logoutClient, req); err != nil {
					return errors.New("gRPC logout failed")
				}
				cmd.Println("Logout successful")
				return nil
			}

			return errors.New("no HTTP or gRPC client available for logout command")
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Authentication token")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file")

	root.AddCommand(cmd)
}
