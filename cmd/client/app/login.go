package app

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// newLoginCommand создаёт новую команду cobra для аутентификации пользователя.
func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login --server-url <url> --username <username> --password <password>",
		Short: "Authenticate an existing user",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, creds, err := parseLoginFlags(cmd)
			if err != nil {
				return err
			}

			app, err := newLoginService(config)
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

			err = app.Login(ctx, creds)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringP("server-url", "s", "", "Server URL")
	cmd.Flags().StringP("username", "u", "", "Username")
	cmd.Flags().StringP("password", "p", "", "User password")
	cmd.Flags().String("hmac-key", "", "HMAC key")
	cmd.Flags().String("rsa-public-key", "", "Path to RSA public key")

	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("server-url")

	return cmd
}

// parseLoginFlags парсит флаги команды login и возвращает конфигурацию клиента,
// учётные данные пользователя и ошибку (если есть).
func parseLoginFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Credentials, error) {
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")
	serverURL, _ := cmd.Flags().GetString("server-url")
	hmacKey, _ := cmd.Flags().GetString("hmac-key")
	rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa-public-key")

	creds := models.NewCredentials(
		models.WithUsername(username),
		models.WithPassword(password),
	)

	config, err := configs.NewClientConfig(
		configs.WithClient(serverURL),
		configs.WithHMACEncoder(hmacKey),
		configs.WithRSAEncoder(rsaPublicKeyPath),
	)
	if err != nil {
		return nil, nil, err
	}

	return config, creds, nil
}

// newLoginService создаёт сервис аутентификации пользователя в зависимости от типа клиента.
func newLoginService(config *configs.ClientConfig) (services.Loginer, error) {
	svc := services.NewLoginContextService()

	switch {
	case config.HTTPClient != nil:
		httpLogin := services.NewHTTPLoginService(config.HTTPClient)
		svc.SetContext(httpLogin)
	case config.GRPCClient != nil:
		grpcClient := pb.NewLoginServiceClient(config.GRPCClient)
		grpcLogin := services.NewGRPCLoginService(grpcClient)
		svc.SetContext(grpcLogin)
	default:
		return nil, errors.New("unsupported server scheme")
	}

	return svc, nil
}
