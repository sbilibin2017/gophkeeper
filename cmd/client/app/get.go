package app

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/spf13/cobra"
)

// newGetLoginPasswordCommand creates a Cobra command to retrieve a login-password secret.
// It supports interactive input mode and uses HTTP client to fetch the secret from the server.
// newGetLoginPasswordCommand creates a Cobra command to retrieve a login-password secret as JSON.
func newGetLoginPasswordCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-login-password",
		Short: "Get secret with login and password",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, secretID, token, err := parseGetSecretFlags(cmd)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			secret, err := services.GetHTTPLoginPassword(ctx, config.HTTPClient, token, secretID)
			if err != nil {
				return fmt.Errorf("failed to get login-password secret: %w", err)
			}

			data, err := json.MarshalIndent(secret, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal secret to JSON: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().String("token", "", "Authentication token")
	cmd.Flags().String("server-url", "", "Server URL")

	cmd.Flags().String("secret_id", "", "ID of the secret to retrieve")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")

	return cmd
}

// parseGetSecretFlags parses flags for secret retrieval commands.
// If interactive mode is enabled, prompts user for secret_id and token via stdin.
// Returns a configured ClientConfig, secret ID, token, or an error.
func parseGetSecretFlags(cmd *cobra.Command) (*configs.ClientConfig, string, string, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")
	secretID, _ := cmd.Flags().GetString("secret_id")
	token, _ := cmd.Flags().GetString("token")
	serverURL, _ := cmd.Flags().GetString("server-url")

	if interactive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter secret_id: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, "", "", fmt.Errorf("input error")
		}
		secretID = strings.TrimSpace(input)

		fmt.Print("Enter token (optional): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, "", "", fmt.Errorf("input error")
		}
		token = strings.TrimSpace(input)

		// Если не был передан сервер, спросим и его
		if serverURL == "" {
			fmt.Print("Enter server URL (optional): ")
			input, err = reader.ReadString('\n')
			if err != nil {
				return nil, "", "", fmt.Errorf("input error")
			}
			serverURL = strings.TrimSpace(input)
		}
	}

	if secretID == "" {
		return nil, "", "", fmt.Errorf("secret_id required")
	}

	config, err := configs.NewClientConfig(
		configs.WithDB(),
		configs.WithHTTPClient(os.Getenv("GOPHKEEPER_SERVER_URL")),
		configs.WithHTTPClient(serverURL),
		configs.WithToken(token),
	)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to create client config: %w", err)
	}

	return config, secretID, token, nil
}

// newGetTextCommand creates a Cobra command to retrieve a text secret.
func newGetTextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-text",
		Short: "Get text secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, secretID, token, err := parseGetSecretFlags(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			secret, err := services.GetHTTPText(ctx, config.HTTPClient, token, secretID)
			if err != nil {
				return fmt.Errorf("failed to get text secret: %w", err)
			}

			data, err := json.MarshalIndent(secret, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal secret to JSON: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().String("token", "", "Authentication token")
	cmd.Flags().String("server-url", "", "Server URL")

	cmd.Flags().String("secret_id", "", "ID of the secret to retrieve")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")

	return cmd
}

// newGetBinaryCommand creates a Cobra command to retrieve a binary secret.
func newGetBinaryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-binary",
		Short: "Get binary secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, secretID, token, err := parseGetSecretFlags(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			secret, err := services.GetHTTPBinary(ctx, config.HTTPClient, token, secretID)
			if err != nil {
				return fmt.Errorf("failed to get binary secret: %w", err)
			}

			data, err := json.MarshalIndent(secret, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal secret to JSON: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().String("token", "", "Authentication token")
	cmd.Flags().String("server-url", "", "Server URL")

	cmd.Flags().String("secret_id", "", "ID of the secret to retrieve")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")

	return cmd
}

// newGetCardCommand creates a Cobra command to retrieve a card secret.
func newGetCardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-card",
		Short: "Get card secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, secretID, token, err := parseGetSecretFlags(cmd)
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			secret, err := services.GetHTTPCard(ctx, config.HTTPClient, token, secretID)
			if err != nil {
				return fmt.Errorf("failed to get card secret: %w", err)
			}

			data, err := json.MarshalIndent(secret, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal secret to JSON: %w", err)
			}

			fmt.Println(string(data))
			return nil
		},
	}

	cmd.Flags().String("token", "", "Authentication token")
	cmd.Flags().String("server-url", "", "Server URL")

	cmd.Flags().String("secret_id", "", "ID of the secret to retrieve")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")

	return cmd
}
