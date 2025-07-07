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

// newRegisterCommand создаёт новую команду cobra для регистрации пользователя.
// Команда принимает параметры сервера, имя пользователя и пароль, а также опциональные ключи для HMAC и RSA.
func newRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register --server-url <url> --username <username> --password <password>",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, creds, err := parseRegisterFlags(cmd)
			if err != nil {
				return err
			}

			app, err := newRegisterService(config)
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

			err = app.Register(ctx, creds)
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

// parseRegisterFlags парсит флаги команды регистрации и возвращает конфигурацию клиента,
// учётные данные пользователя и возможную ошибку.
func parseRegisterFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Credentials, error) {
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

// newRegisterService создаёт и возвращает сервис регистрации на основе переданной конфигурации клиента.
// В зависимости от типа клиента (HTTP или gRPC) создаётся соответствующий сервис.
// Возвращает ошибку, если схема сервера не поддерживается.
func newRegisterService(config *configs.ClientConfig) (services.Registerer, error) {
	svc := services.NewRegisterContextService()

	switch {
	case config.HTTPClient != nil:
		httpRegister := services.NewHTTPRegisterService(config.HTTPClient)
		svc.SetContext(httpRegister)
	case config.GRPCClient != nil:
		grpcClient := pb.NewRegisterServiceClient(config.GRPCClient)
		grpcRegister := services.NewGRPCRegisterService(grpcClient)
		svc.SetContext(grpcRegister)
	default:
		return nil, errors.New("unsupported server scheme")
	}

	return svc, nil
}
