package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/protocol"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/spf13/cobra"
)

func newRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register -server-url <url> -username <username> -password <password> [-rsa-public-key-path <path>] [-hmac-key <key>]",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, creds, err := parseFlags(cmd)
			if err != nil {
				return err
			}

			svc, err := newRegisterService(
				config.ServerURL,
				config.RSAPublicKeyPath,
				config.HMACKey,
			)
			if err != nil {
				return err
			}

			if err := svc.Register(context.Background(), creds); err != nil {
				return err
			}

			// Выводим сообщение об успешной регистрации, чтобы тест мог проверить
			fmt.Fprintf(cmd.OutOrStdout(), "User %q registered successfully\n", creds.Username)
			return nil
		},
	}

	cmd.Flags().String("server-url", "https://localhost:8000", "Server URL")
	cmd.Flags().String("username", "", "Username")
	cmd.Flags().String("password", "", "User password")
	cmd.Flags().String("rsa-public-key-path", "", "Path to RSA public key file (optional)")
	cmd.Flags().String("hmac-key", "", "HMAC key (optional)")

	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

func parseFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Credentials, error) {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return nil, nil, err
	}
	if serverURL == "" {
		return nil, nil, errors.New("server-url is required")
	}

	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return nil, nil, err
	}
	if username == "" {
		return nil, nil, errors.New("username is required")
	}

	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return nil, nil, err
	}
	if password == "" {
		return nil, nil, errors.New("password is required")
	}

	rsaPublicKeyPath, err := cmd.Flags().GetString("rsa-public-key-path")
	if err != nil {
		return nil, nil, err
	}

	hmacKey, err := cmd.Flags().GetString("hmac-key")
	if err != nil {
		return nil, nil, err
	}

	config := configs.NewClientConfig(
		configs.WithServerURL(serverURL),
		configs.WithRSAPublicKeyPath(rsaPublicKeyPath),
		configs.WithHMACKey(hmacKey),
	)

	creds := models.NewCredentials(
		models.WithUsername(username),
		models.WithPassword(password),
	)

	return config, creds, nil
}

func newRegisterService(serverURL, rsaPublicKeyPath, hmacKey string) (*services.RegisterService, error) {
	service := services.NewRegisterService()

	p, err := protocol.GetProtocol(serverURL)
	if err != nil {
		return nil, err
	}

	switch p {
	case protocol.HTTP, protocol.HTTPS:
		svc := services.NewRegisterHTTPService(
			services.WithHTTPServerURL(serverURL),
			services.WithHTTPPublicKeyPath(rsaPublicKeyPath),
		)
		service.SetContext(svc)
	case protocol.GRPC:
		svc := services.NewRegisterGRPCService(
			services.WithGRPCServerURL(serverURL),
			services.WithGRPCPublicKeyPath(rsaPublicKeyPath),
		)
		service.SetContext(svc)
	default:
		return nil, errors.New("unknown server protocol")
	}

	return service, nil
}
