package app

import (
	"context"
	"errors"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/protocol"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/spf13/cobra"
)

// newRegisterCommand creates a cobra.Command for registering a new user.
// The command requires flags: server-url, username, password,
// and optional flags: rsa-public-key-path and hmac-key.
// On successful registration, it prints a success message.
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

			return nil
		},
	}

	// Define command flags
	cmd.Flags().String("server-url", "https://localhost:8000", "Server URL")
	cmd.Flags().String("username", "", "Username")
	cmd.Flags().String("password", "", "User password")
	cmd.Flags().String("rsa-public-key-path", "", "Path to RSA public key file (optional)")
	cmd.Flags().String("hmac-key", "", "HMAC key (optional)")

	// Mark required flags
	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

// parseFlags extracts and validates flag values from the command,
// returning client configuration and user credentials.
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

	// Create client configuration
	config := configs.NewClientConfig(
		configs.WithServerURL(serverURL),
		configs.WithRSAPublicKeyPath(rsaPublicKeyPath),
		configs.WithHMACKey(hmacKey),
	)

	// Create user credentials
	creds := models.NewCredentials(
		models.WithUsername(username),
		models.WithPassword(password),
	)

	return config, creds, nil
}

// newRegisterService creates a registration service depending on the server protocol.
// Supports HTTP(S) and gRPC.
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
