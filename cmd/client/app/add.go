package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
)

func newAddLoginPasswordCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-login-password",
		Short: "Add a login-password secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server_url")
			hmacKey, _ := cmd.Flags().GetString("hmac_key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa_public_key")
			interactive, _ := cmd.Flags().GetBool("interactive")

			secretID, _ := cmd.Flags().GetString("secret_id")
			login, _ := cmd.Flags().GetString("login")
			password, _ := cmd.Flags().GetString("password")
			metaFlag, _ := cmd.Flags().GetStringToString("meta")

			// If interactive, read from stdin instead
			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter secret_id: ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				secretID = strings.TrimSpace(input)

				fmt.Print("Enter login: ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				login = strings.TrimSpace(input)

				fmt.Print("Enter password: ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				password = strings.TrimSpace(input)

				fmt.Print("Enter meta (key=value pairs separated by commas, optional): ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				metaStr := strings.TrimSpace(input)
				metaFlag = parseAddMetaString(metaStr)
			}

			if secretID == "" {
				return fmt.Errorf("secret_id cannot be empty")
			}
			if login == "" {
				return fmt.Errorf("login cannot be empty")
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

			req := models.NewLoginPassword(
				models.WithLoginPasswordSecretID(secretID),
				models.WithLoginPasswordLogin(login),
				models.WithLoginPasswordPassword(password),
				models.WithLoginPasswordMeta(metaFlag),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if config.HTTPClient != nil {
				err = services.AddLoginPasswordHTTP(
					ctx,
					req,
					services.WithAddLoginPasswordHTTPClient(config.HTTPClient),
					services.WithAddLoginPasswordHTTPEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				return nil
			}

			if config.GRPCClient != nil {
				client := pb.NewAddLoginPasswordServiceClient(config.GRPCClient)
				err := services.AddLoginPasswordGRPC(
					ctx,
					req,
					services.WithAddLoginPasswordGRPCClient(client),
					services.WithAddLoginPasswordGRPCEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				return nil
			}

			return fmt.Errorf("no client configured for adding login-password secret")
		},
	}

	cmd.Flags().String("server_url", "", "Server URL")
	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().String("secret_id", "", "ID of the secret")
	cmd.Flags().String("login", "", "Login username to store")
	cmd.Flags().String("password", "", "Password to store")
	cmd.Flags().StringToString("meta", nil, "Optional metadata key=value pairs")

	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("server_url")
	_ = cmd.MarkFlagRequired("secret_id")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}

func newAddTextSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-text",
		Short: "Add a text secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server_url")
			hmacKey, _ := cmd.Flags().GetString("hmac_key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa_public_key")
			interactive, _ := cmd.Flags().GetBool("interactive")

			secretID, _ := cmd.Flags().GetString("secret_id")
			content, _ := cmd.Flags().GetString("content")
			metaFlag, _ := cmd.Flags().GetStringToString("meta")

			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter secret_id: ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				secretID = strings.TrimSpace(input)

				fmt.Print("Enter content: ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				content = strings.TrimSpace(input)

				fmt.Print("Enter meta (key=value pairs separated by commas, optional): ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				metaStr := strings.TrimSpace(input)
				metaFlag = parseAddMetaString(metaStr)
			}

			if secretID == "" {
				return fmt.Errorf("secret_id cannot be empty")
			}
			if content == "" {
				return fmt.Errorf("content cannot be empty")
			}

			config, err := configs.NewClientConfig(
				configs.WithClient(serverURL),
				configs.WithHMACEncoder(hmacKey),
				configs.WithRSAEncoder(rsaPublicKeyPath),
			)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			req := models.NewText(
				models.WithTextSecretID(secretID),
				models.WithTextContent(content),
				models.WithTextMeta(metaFlag),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if config.HTTPClient != nil {
				err = services.AddTextSecretHTTP(
					ctx,
					req,
					services.WithAddTextSecretHTTPClient(config.HTTPClient),
					services.WithAddTextSecretHTTPEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				return nil
			}

			if config.GRPCClient != nil {
				client := pb.NewAddTextServiceClient(config.GRPCClient)
				err = services.AddTextSecretGRPC(
					ctx,
					req,
					services.WithAddTextSecretGRPCClient(client),
					services.WithAddTextSecretGRPCEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				return nil
			}

			return fmt.Errorf("no client configured for adding text secret")
		},
	}

	cmd.Flags().String("server_url", "", "Server URL")
	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().String("secret_id", "", "ID of the text secret")
	cmd.Flags().String("content", "", "Text content to store")
	cmd.Flags().StringToString("meta", nil, "Optional metadata key=value pairs")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("server_url")
	_ = cmd.MarkFlagRequired("secret_id")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}

func newAddBinarySecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Add a binary secret (e.g., file data)",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server_url")
			hmacKey, _ := cmd.Flags().GetString("hmac_key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa_public_key")
			interactive, _ := cmd.Flags().GetBool("interactive")

			secretID, _ := cmd.Flags().GetString("secret_id")
			filePath, _ := cmd.Flags().GetString("file")
			meta, _ := cmd.Flags().GetStringToString("meta")

			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter secret_id: ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				secretID = strings.TrimSpace(input)

				fmt.Print("Enter file path: ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				filePath = strings.TrimSpace(input)

				fmt.Print("Enter meta (key=value pairs separated by commas, optional): ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				metaStr := strings.TrimSpace(input)
				meta = parseAddMetaString(metaStr)
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				cmd.PrintErrln("Failed to read file:", err)
				return err
			}

			config, err := configs.NewClientConfig(
				configs.WithClient(serverURL),
				configs.WithHMACEncoder(hmacKey),
				configs.WithRSAEncoder(rsaPublicKeyPath),
			)
			if err != nil {
				cmd.PrintErrln("Failed to create client config:", err)
				return err
			}

			secret := models.NewBinary(
				models.WithBinarySecretID(secretID),
				models.WithBinaryData(data),
				models.WithBinaryMeta(meta),
				models.WithBinaryUpdatedAt(time.Now()),
			)

			ctx := context.Background()

			if config.HTTPClient != nil {
				err = services.AddBinarySecretHTTP(
					ctx,
					secret,
					services.WithAddBinarySecretHTTPClient(config.HTTPClient),
					services.WithAddBinarySecretHTTPEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				cmd.Println("Binary secret added via HTTP")
				return nil
			}

			if config.GRPCClient != nil {
				client := pb.NewAddBinaryServiceClient(config.GRPCClient)
				err = services.AddBinarySecretGRPC(
					ctx,
					secret,
					services.WithAddBinarySecretGRPCClient(client),
					services.WithAddBinarySecretGRPCEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				cmd.Println("Binary secret added via gRPC")
				return nil
			}

			return fmt.Errorf("no client configured for adding binary secret")
		},
	}

	cmd.Flags().String("server_url", "", "Server URL")
	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().String("secret_id", "", "ID of the binary secret")
	cmd.Flags().String("file", "", "Path to the file to store")
	cmd.Flags().StringToString("meta", nil, "Optional metadata (key=value pairs separated by commas)")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	cmd.MarkFlagRequired("server_url")
	cmd.MarkFlagRequired("secret_id")
	cmd.MarkFlagRequired("file")

	return cmd
}

func newAddCardSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-card",
		Short: "Add a card secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, _ := cmd.Flags().GetString("server_url")
			hmacKey, _ := cmd.Flags().GetString("hmac_key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa_public_key")
			interactive, _ := cmd.Flags().GetBool("interactive")

			secretID, _ := cmd.Flags().GetString("secret_id")
			number, _ := cmd.Flags().GetString("number")
			holder, _ := cmd.Flags().GetString("holder")
			expMonth, _ := cmd.Flags().GetInt("exp_month")
			expYear, _ := cmd.Flags().GetInt("exp_year")
			cvv, _ := cmd.Flags().GetString("cvv")
			meta, _ := cmd.Flags().GetStringToString("meta")

			if interactive {
				reader := bufio.NewReader(os.Stdin)

				fmt.Print("Enter secret_id: ")
				input, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				secretID = strings.TrimSpace(input)

				fmt.Print("Enter card number: ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				number = strings.TrimSpace(input)

				fmt.Print("Enter cardholder name: ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				holder = strings.TrimSpace(input)

				fmt.Print("Enter expiration month (1-12): ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				em, err := strconv.Atoi(strings.TrimSpace(input))
				if err != nil {
					return fmt.Errorf("invalid expiration month: %v", err)
				}
				expMonth = em

				fmt.Print("Enter expiration year (e.g. 2025): ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				ey, err := strconv.Atoi(strings.TrimSpace(input))
				if err != nil {
					return fmt.Errorf("invalid expiration year: %v", err)
				}
				expYear = ey

				fmt.Print("Enter CVV: ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				cvv = strings.TrimSpace(input)

				fmt.Print("Enter meta (key=value pairs separated by commas, optional): ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				metaStr := strings.TrimSpace(input)
				meta = parseAddMetaString(metaStr)
			}

			config, err := configs.NewClientConfig(
				configs.WithClient(serverURL),
				configs.WithHMACEncoder(hmacKey),
				configs.WithRSAEncoder(rsaPublicKeyPath),
			)
			if err != nil {
				cmd.PrintErrln("Failed to create client config:", err)
				return err
			}

			secret := models.NewCard(
				models.WithCardSecretID(secretID),
				models.WithCardNumber(number),
				models.WithCardHolder(holder),
				models.WithCardExpMonth(expMonth),
				models.WithCardExpYear(expYear),
				models.WithCardCVV(cvv),
				models.WithCardMeta(meta),
				models.WithCardUpdatedAt(time.Now()),
			)

			ctx := context.Background()

			if config.HTTPClient != nil {
				err = services.AddCardSecretHTTP(
					ctx,
					secret,
					services.WithAddCardSecretHTTPClient(config.HTTPClient),
					services.WithAddCardSecretHTTPEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				cmd.Println("Card secret added via HTTP")
				return nil
			}

			if config.GRPCClient != nil {
				client := pb.NewAddCardServiceClient(config.GRPCClient)
				err = services.AddCardSecretGRPC(
					ctx,
					secret,
					services.WithAddCardSecretGRPCClient(client),
					services.WithAddCardSecretGRPCEncoders(config.Encoders),
				)
				if err != nil {
					return err
				}
				cmd.Println("Card secret added via gRPC")
				return nil
			}

			cmd.Println("No client configured (HTTP or gRPC)")
			return fmt.Errorf("no client configured")
		},
	}

	cmd.Flags().String("server_url", "", "Server URL")
	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().String("secret_id", "", "ID of the card secret")
	cmd.Flags().String("number", "", "Card number")
	cmd.Flags().String("holder", "", "Cardholder name")
	cmd.Flags().Int("exp_month", 0, "Expiration month (1-12)")
	cmd.Flags().Int("exp_year", 0, "Expiration year (e.g. 2025)")
	cmd.Flags().String("cvv", "", "CVV code")
	cmd.Flags().StringToString("meta", nil, "Optional metadata (key=value pairs separated by commas)")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	cmd.MarkFlagRequired("server_url")
	cmd.MarkFlagRequired("secret_id")
	cmd.MarkFlagRequired("number")
	cmd.MarkFlagRequired("holder")
	cmd.MarkFlagRequired("exp_month")
	cmd.MarkFlagRequired("exp_year")
	cmd.MarkFlagRequired("cvv")

	return cmd
}

// parseAddMetaString converts a comma-separated key=value string into a map[string]string
func parseAddMetaString(metaStr string) map[string]string {
	meta := make(map[string]string)
	if metaStr == "" {
		return meta
	}
	pairs := strings.Split(metaStr, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			meta[kv[0]] = kv[1]
		}
	}
	return meta
}
