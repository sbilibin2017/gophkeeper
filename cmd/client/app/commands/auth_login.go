package commands

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/validation"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

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
		Long:  "Authenticate user by username and password.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := validation.ValidateString("username", username); err != nil {
				return err
			}
			if err := validation.ValidateString("password", password); err != nil {
				return err
			}

			cfg, err := config.NewClientConfig(authURL, tlsClientCert, tlsClientKey)
			if err != nil {
				return err
			}

			if err := client.CreateBinaryRequestTable(ctx, cfg.DB); err != nil {
				return err
			}
			if err := client.CreateTextRequestTable(ctx, cfg.DB); err != nil {
				return err
			}
			if err := client.CreateUsernamePasswordRequestTable(ctx, cfg.DB); err != nil {
				return err
			}
			if err := client.CreateBankCardRequestTable(ctx, cfg.DB); err != nil {
				return err
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
				cmd.Println("HTTP login failed, trying gRPC fallback...")
			}

			if cfg.GRPCClient != nil {
				loginClient := pb.NewLoginServiceClient(cfg.GRPCClient)
				resp, err := client.LoginGRPC(ctx, loginClient, req)
				if err != nil {
					return errors.New("gRPC login failed")
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

	root.AddCommand(cmd)
}
