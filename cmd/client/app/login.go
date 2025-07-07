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

	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

// parseLoginFlags parses the flags of the login command and returns the client configuration,
// user credentials, and an error (if any).
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

// Loginer describes the interface for the user authentication service.
type Loginer interface {
	// Login performs user authentication with the specified credentials.
	Login(ctx context.Context, creds *models.Credentials) error
}

// newLoginService creates a user authentication service depending on the client type.
func newLoginService(config *configs.ClientConfig) (Loginer, error) {
	switch {
	case config.HTTPClient != nil:
		svc := services.NewHTTPLoginService(config.HTTPClient)
		return svc, nil
	case config.GRPCClient != nil:
		grpcClient := pb.NewLoginServiceClient(config.GRPCClient)
		svc := services.NewGRPCLoginService(grpcClient)
		return svc, nil
	default:
		return nil, errors.New("unsupported server scheme")
	}
}
