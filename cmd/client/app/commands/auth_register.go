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
		Long:  "Register a new user account by providing a username and password.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			if err := validation.ValidateRegisterUsername(username); err != nil {
				return err
			}
			if err := validation.ValidateRegisterPassword(password); err != nil {
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

			req := &models.RegisterRequest{
				Username: username,
				Password: password,
			}

			if cfg.HTTPClient != nil {
				if err := client.RegisterHTTP(ctx, cfg.HTTPClient, req); err != nil {
					return err
				}
			}

			if cfg.GRPCClient != nil {
				registerClient := pb.NewRegisterServiceClient(cfg.GRPCClient)
				if err := client.RegisterGRPC(ctx, registerClient, req); err != nil {
					return err
				}
				return nil
			}

			return errors.New("no HTTP or gRPC client available for register command")
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&username, "username", "", "Username for registration")
	flags.StringVar(&password, "password", "", "Password for registration")
	flags.StringVar(&authURL, "auth-url", "", "Authentication service URL")
	flags.StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file")
	flags.StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS key file")

	root.AddCommand(cmd)
}
