package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
)

// newLoginCommand creates the Cobra command for user login.
func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login an existing user",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server-url")
			hmacKey, _ := cmd.Flags().GetString("hmac-key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa-public-key")
			interactive, _ := cmd.Flags().GetBool("interactive")

			username, _ := cmd.Flags().GetString("username")
			password, _ := cmd.Flags().GetString("password")

			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter username: ")
				uInput, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				username = strings.TrimSpace(uInput)

				fmt.Print("Enter password: ")
				pInput, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				password = strings.TrimSpace(pInput)
			}

			if username == "" {
				return fmt.Errorf("username cannot be empty")
			}
			if password == "" {
				return fmt.Errorf("password cannot be empty")
			}

			config, err := configs.NewClientConfig(
				configs.WithClient(serverURL),
				configs.WithHMACEncoder(hmacKey),
				configs.WithRSAEncoder(rsaPublicKeyPath),
			)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			req := models.NewUser(
				models.WithUsername(username),
				models.WithPassword(password),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if config.HTTPClient != nil {
				// Use functional options for RegisterHTTP
				err = services.LoginHTTP(
					ctx,
					req,
					services.WithLoginHTTPClient(config.HTTPClient),
					services.WithLoginHTTPEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				return nil
			}

			if config.GRPCClient != nil {
				client := pb.NewLoginServiceClient(config.GRPCClient)

				err = services.LoginGRPC(
					ctx,
					req,
					services.WithLoginGRPCClient(client),
					services.WithLoginGRPCEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				return nil
			}

			return fmt.Errorf("no client configured for registration")
		},
	}

	cmd.Flags().String("server-url", "", "Server URL")
	cmd.Flags().String("hmac-key", "", "HMAC key")
	cmd.Flags().String("rsa-public-key", "", "Path to RSA public key")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")
	cmd.Flags().String("username", "", "Username")
	cmd.Flags().String("password", "", "User password")

	_ = cmd.MarkFlagRequired("server-url")

	return cmd
}
