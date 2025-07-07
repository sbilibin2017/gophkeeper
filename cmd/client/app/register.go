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

// newRegisterCommand creates a new cobra command for user registration.
// The command accepts server URL, username, and password parameters,
// as well as optional keys for HMAC and RSA.
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

	// Move the server-url flag to the first position
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

// parseRegisterFlags parses the flags of the register command and returns
// the client configuration, user credentials, and any error.
func parseRegisterFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Credentials, error) {
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
		return nil, nil, err
	}

	creds := models.NewCredentials(
		models.WithUsername(username),
		models.WithPassword(password),
	)

	return config, creds, nil
}

// Registerer describes the user registration service interface.
type Registerer interface {
	// Register performs user registration with the specified credentials.
	Register(ctx context.Context, creds *models.Credentials) error
}

// newRegisterService creates and returns a registration service based on the provided client configuration.
// Depending on the client type (HTTP or gRPC), it creates the corresponding service.
// Returns an error if the server scheme is not supported.
func newRegisterService(config *configs.ClientConfig) (Registerer, error) {
	switch {
	case config.HTTPClient != nil:
		svc := services.NewHTTPRegisterService(config.HTTPClient)
		return svc, nil
	case config.GRPCClient != nil:
		grpcClient := pb.NewRegisterServiceClient(config.GRPCClient)
		svc := services.NewGRPCRegisterService(grpcClient)
		return svc, nil
	default:
		return nil, errors.New("unsupported server scheme")
	}
}
