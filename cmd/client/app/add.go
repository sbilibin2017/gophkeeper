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

	"github.com/spf13/cobra"
)

func newAddLoginPasswordCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-login-password",
		Short: "Add a login-password secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			hmacKey, _ := cmd.Flags().GetString("hmac_key")
			rsaPublicKeyPath, _ := cmd.Flags().GetString("rsa_public_key")
			interactive, _ := cmd.Flags().GetBool("interactive")

			secretID, _ := cmd.Flags().GetString("secret_id")
			login, _ := cmd.Flags().GetString("login")
			password, _ := cmd.Flags().GetString("password")
			metaFlag, _ := cmd.Flags().GetStringToString("meta")

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
				metaFlag = parseAddMetaString(strings.TrimSpace(input))
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
				configs.WithDB(),
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

			err = services.AddLoginPassword(ctx, req,
				services.WithAddLoginPasswordEncoders(config.Encoders),
				services.WithAddLoginPasswordDB(config.DB),
			)
			if err != nil {
				return fmt.Errorf("failed to add login-password secret: %w", err)
			}

			fmt.Println("Login-password secret added successfully")
			return nil
		},
	}

	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().String("secret_id", "", "ID of the secret")
	cmd.Flags().String("login", "", "Login username to store")
	cmd.Flags().String("password", "", "Password to store")
	cmd.Flags().StringToString("meta", nil, "Optional metadata key=value pairs")

	cmd.Flags().Bool("interactive", false, "Enable interactive input")

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
				metaFlag = parseAddMetaString(strings.TrimSpace(input))
			}

			if secretID == "" {
				return fmt.Errorf("secret_id cannot be empty")
			}
			if content == "" {
				return fmt.Errorf("content cannot be empty")
			}

			config, err := configs.NewClientConfig(
				configs.WithDB(),
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

			err = services.AddText(ctx, req,
				services.WithAddTextEncoders(config.Encoders),
				services.WithAddTextDB(config.DB),
			)
			if err != nil {
				return fmt.Errorf("failed to add text secret: %w", err)
			}

			fmt.Println("Text secret added successfully")
			return nil
		},
	}

	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().String("secret_id", "", "ID of the text secret")
	cmd.Flags().String("content", "", "Text content to store")
	cmd.Flags().StringToString("meta", nil, "Optional metadata key=value pairs")

	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}

func newAddBinarySecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Add a binary secret (e.g., file data)",
		RunE: func(cmd *cobra.Command, args []string) error {
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
				meta = parseAddMetaString(strings.TrimSpace(input))
			}

			if secretID == "" {
				return fmt.Errorf("secret_id cannot be empty")
			}
			if filePath == "" {
				return fmt.Errorf("file path cannot be empty")
			}

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			config, err := configs.NewClientConfig(
				configs.WithDB(),
				configs.WithHMACEncoder(hmacKey),
				configs.WithRSAEncoder(rsaPublicKeyPath),
			)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			req := models.NewBinary(
				models.WithBinarySecretID(secretID),
				models.WithBinaryData(data),
				models.WithBinaryMeta(meta),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = services.AddBinary(ctx, req,
				services.WithAddBinaryEncoders(config.Encoders),
				services.WithAddBinaryDB(config.DB),
			)
			if err != nil {
				return fmt.Errorf("failed to add binary secret: %w", err)
			}

			fmt.Println("Binary secret added successfully")
			return nil
		},
	}

	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().String("secret_id", "", "ID of the binary secret")
	cmd.Flags().String("file", "", "Path to the file containing binary data")
	cmd.Flags().StringToString("meta", nil, "Optional metadata key=value pairs")

	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func newAddCardSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-card",
		Short: "Add a card secret",
		RunE: func(cmd *cobra.Command, args []string) error {
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

				fmt.Print("Enter card holder name: ")
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
				m, err := strconv.Atoi(strings.TrimSpace(input))
				if err != nil || m < 1 || m > 12 {
					return fmt.Errorf("invalid expiration month")
				}
				expMonth = m

				fmt.Print("Enter expiration year (e.g. 2025): ")
				input, err = reader.ReadString('\n')
				if err != nil {
					return err
				}
				y, err := strconv.Atoi(strings.TrimSpace(input))
				if err != nil || y < time.Now().Year() {
					return fmt.Errorf("invalid expiration year")
				}
				expYear = y

				fmt.Print("Enter CVV code: ")
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
				meta = parseAddMetaString(strings.TrimSpace(input))
			}

			if secretID == "" {
				return fmt.Errorf("secret_id cannot be empty")
			}
			if number == "" {
				return fmt.Errorf("card number cannot be empty")
			}
			if holder == "" {
				return fmt.Errorf("card holder cannot be empty")
			}
			if expMonth < 1 || expMonth > 12 {
				return fmt.Errorf("expiration month must be between 1 and 12")
			}
			if expYear < time.Now().Year() {
				return fmt.Errorf("expiration year must not be in the past")
			}
			if cvv == "" {
				return fmt.Errorf("cvv cannot be empty")
			}

			config, err := configs.NewClientConfig(
				configs.WithDB(),
				configs.WithHMACEncoder(hmacKey),
				configs.WithRSAEncoder(rsaPublicKeyPath),
			)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			req := models.NewCard(
				models.WithCardSecretID(secretID),
				models.WithCardNumber(number),
				models.WithCardHolder(holder),
				models.WithCardExpMonth(expMonth),
				models.WithCardExpYear(expYear),
				models.WithCardCVV(cvv),
				models.WithCardMeta(meta),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = services.AddCard(ctx, req,
				services.WithAddCardEncoders(config.Encoders),
				services.WithAddCardDB(config.DB),
			)
			if err != nil {
				return fmt.Errorf("failed to add card secret: %w", err)
			}

			fmt.Println("Card secret added successfully")
			return nil
		},
	}

	cmd.Flags().String("hmac_key", "", "HMAC key")
	cmd.Flags().String("rsa_public_key", "", "Path to RSA public key")

	cmd.Flags().String("secret_id", "", "ID of the card secret")
	cmd.Flags().String("number", "", "Card number")
	cmd.Flags().String("holder", "", "Cardholder name")
	cmd.Flags().Int("exp_month", 0, "Expiration month (1-12)")
	cmd.Flags().Int("exp_year", 0, "Expiration year (e.g. 2025)")
	cmd.Flags().String("cvv", "", "CVV code")
	cmd.Flags().StringToString("meta", nil, "Optional metadata key=value pairs")

	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("holder")
	_ = cmd.MarkFlagRequired("exp_month")
	_ = cmd.MarkFlagRequired("exp_year")
	_ = cmd.MarkFlagRequired("cvv")

	return cmd
}

func parseAddMetaString(input string) map[string]string {
	meta := make(map[string]string)
	if input == "" {
		return meta
	}
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(kv) == 2 {
			meta[kv[0]] = kv[1]
		}
	}
	return meta
}
