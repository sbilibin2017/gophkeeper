package app

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/services"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func newRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register --server-url <url> --username <username> --password <password>",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")
			serverURL, _ := cmd.Flags().GetString("server-url")
			hmacKey, _ := cmd.Flags().GetString("hmac-key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa-public-key")

			config, err := configs.NewClientConfig(
				configs.WithClient(serverURL),
				configs.WithHMACEncoder(hmacKey),
				configs.WithRSAEncoder(rsaPublicKeyPath),
			)
			if err != nil {
				return err
			}

			defer func() {
				if config.GRPCClient != nil {
					config.GRPCClient.Close()
				}
			}()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			switch {
			case config.HTTPClient != nil:
				return services.RegisterHTTP(
					ctx,
					config.HTTPClient.SetBaseURL("/register/"),
					config.HMACEncoder,
					config.RSAEncoder,
					username, password,
				)

			case config.GRPCClient != nil:
				client := pb.NewRegisterServiceClient(config.GRPCClient)
				return services.RegisterGRPC(
					ctx,
					client,
					config.HMACEncoder,
					config.RSAEncoder,
					username,
					password,
				)

			default:
				return errors.New("unsupported server scheme")
			}
		},
	}

	cmd.Flags().StringP("server-url", "s", "", "Server URL")
	cmd.Flags().StringP("username", "u", "", "Username")
	cmd.Flags().StringP("password", "p", "", "User password")
	cmd.Flags().String("hmac-key", "", "HMAC key")
	cmd.Flags().String("rsa-public-key", "", "Path to RSA public key")

	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}
