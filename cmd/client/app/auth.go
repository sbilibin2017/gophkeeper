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
	"github.com/spf13/cobra"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// newRegisterCommand creates a Cobra command to register a new user.
// Supports flags --server-url, --interactive, --username and --password.
//
// CLI usage example:
//
//	gophkeeper register --server-url http://localhost:8080 --username user --password secret
func newRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, creds, err := parseRegisterFlags(cmd)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			token, err := runRegisterApp(ctx, config, creds)
			if err != nil {
				return err
			}

			fmt.Println(token)
			return nil
		},
	}

	cmd.Flags().String("server-url", "", "Server URL (http:// or grpc://)")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")
	cmd.Flags().String("username", "", "Username")
	cmd.Flags().String("password", "", "User password")

	_ = cmd.MarkFlagRequired("server-url")

	return cmd
}

// parseRegisterFlags parses flags of the register command and returns client config and user credentials.
// If the --interactive flag is set, it prompts the user for input via stdin.
func parseRegisterFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Credentials, error) {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get 'server-url' flag")
	}

	interactive, _ := cmd.Flags().GetBool("interactive")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	if interactive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter username: ")
		uInput, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read username")
		}
		username = strings.TrimSpace(uInput)

		fmt.Print("Enter password: ")
		pInput, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read password")
		}
		password = strings.TrimSpace(pInput)
	}

	if username == "" {
		return nil, nil, fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return nil, nil, fmt.Errorf("password cannot be empty")
	}

	var opts []configs.ClientConfigOpt

	if strings.HasPrefix(serverURL, "http://") {
		opts = append(opts, configs.WithHTTPClient(serverURL))
	} else if strings.HasPrefix(serverURL, "grpc://") {
		opts = append(opts, configs.WithGRPCClient(serverURL))
	} else {
		return nil, nil, fmt.Errorf("unsupported server URL scheme")
	}

	config, err := configs.NewClientConfig(opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create client config")
	}

	creds := &models.Credentials{
		Username: username,
		Password: password,
	}

	return config, creds, nil
}

// runRegisterApp performs user registration using HTTP or gRPC client,
// depending on the configuration.
func runRegisterApp(ctx context.Context, config *configs.ClientConfig, creds *models.Credentials) (string, error) {
	if config.HTTPClient != nil {
		token, err := services.RegisterHTTP(ctx, config.HTTPClient, creds)
		if err != nil {
			return "", fmt.Errorf("HTTP registration failed")
		}
		return token, nil
	}

	if config.GRPCClient != nil {
		client := pb.NewRegisterServiceClient(config.GRPCClient)
		token, err := services.RegisterGRPC(ctx, client, creds)
		if err != nil {
			return "", fmt.Errorf("gRPC registration failed")
		}
		return token, nil
	}

	return "", fmt.Errorf("no client configured for registration")
}

// newLoginCommand creates a Cobra command to login an existing user.
// Supports flags --server-url, --interactive, --username, --password.
//
// Example:
//
//	gophkeeper login --server-url http://localhost:8080 --username user --password secret
func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login an existing user",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, creds, err := parseLoginFlags(cmd)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			token, err := runLoginApp(ctx, config, creds)
			if err != nil {
				return err
			}

			fmt.Println(token)
			return nil
		},
	}

	cmd.Flags().String("server-url", "", "Server URL (http:// or grpc://)")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")
	cmd.Flags().String("username", "", "Username")
	cmd.Flags().String("password", "", "User password")

	_ = cmd.MarkFlagRequired("server-url")

	return cmd
}

// parseLoginFlags parses flags of the login command and returns client config and user credentials.
// Supports --interactive flag for manual input.
func parseLoginFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Credentials, error) {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get 'server-url' flag")
	}

	interactive, _ := cmd.Flags().GetBool("interactive")
	username, _ := cmd.Flags().GetString("username")
	password, _ := cmd.Flags().GetString("password")

	if interactive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter username: ")
		uInput, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read username")
		}
		username = strings.TrimSpace(uInput)

		fmt.Print("Enter password: ")
		pInput, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read password")
		}
		password = strings.TrimSpace(pInput)
	}

	if username == "" {
		return nil, nil, fmt.Errorf("username cannot be empty")
	}
	if password == "" {
		return nil, nil, fmt.Errorf("password cannot be empty")
	}

	var opts []configs.ClientConfigOpt

	if strings.HasPrefix(serverURL, "http://") {
		opts = append(opts, configs.WithHTTPClient(serverURL))
	} else if strings.HasPrefix(serverURL, "grpc://") {
		opts = append(opts, configs.WithGRPCClient(serverURL))
	} else {
		return nil, nil, fmt.Errorf("unsupported server URL scheme")
	}

	config, err := configs.NewClientConfig(opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create client config")
	}

	creds := &models.Credentials{
		Username: username,
		Password: password,
	}

	return config, creds, nil
}

// runLoginApp performs user login using HTTP or gRPC client,
// depending on the configuration.
func runLoginApp(ctx context.Context, config *configs.ClientConfig, creds *models.Credentials) (string, error) {
	if config.HTTPClient != nil {
		token, err := services.LoginHTTP(ctx, config.HTTPClient, creds)
		if err != nil {
			return "", fmt.Errorf("HTTP login failed")
		}
		return token, nil
	}

	if config.GRPCClient != nil {
		client := pb.NewLoginServiceClient(config.GRPCClient)
		token, err := services.LoginGRPC(ctx, client, creds)
		if err != nil {
			return "", fmt.Errorf("gRPC login failed")
		}
		return token, nil
	}

	return "", fmt.Errorf("no client configured for login")
}
