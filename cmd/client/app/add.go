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

// newAddLoginPasswordCommand creates a Cobra command to add a secret with login and password.
func newAddLoginPasswordCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-login-password",
		Short: "Add secret with login password",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, loginPassword, err := parseAddLoginPasswordFlags(cmd)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = services.AddLoginPassword(ctx, config.DB, loginPassword)
			if err != nil {
				return fmt.Errorf("failed to save secret")
			}

			return nil
		},
	}

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

func parseAddLoginPasswordFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.LoginPassword, error) {
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
			return nil, nil, fmt.Errorf("input error")
		}
		secretID = strings.TrimSpace(input)

		fmt.Print("Enter login: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		login = strings.TrimSpace(input)

		fmt.Print("Enter password: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		password = strings.TrimSpace(input)

		fmt.Print("Enter meta (key=value pairs separated by commas, optional): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		metaFlag = parseAddMetaString(strings.TrimSpace(input))
	}

	if secretID == "" {
		return nil, nil, fmt.Errorf("secret_id required")
	}
	if login == "" {
		return nil, nil, fmt.Errorf("login required")
	}
	if password == "" {
		return nil, nil, fmt.Errorf("password required")
	}

	config, err := configs.NewClientConfig(configs.WithDB())
	if err != nil {
		return nil, nil, fmt.Errorf("db connection error")
	}

	req := &models.LoginPassword{
		SecretID:  secretID,
		Login:     login,
		Password:  password,
		Meta:      metaFlag,
		UpdatedAt: time.Now().UTC(),
	}

	return config, req, nil
}

// newAddTextSecretCommand creates a Cobra command to add a text secret.
func newAddTextSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-text",
		Short: "Add a text secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, textSecret, err := parseAddTextFlags(cmd)
			if err != nil {
				return fmt.Errorf("input error")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = services.AddText(ctx, config.DB, textSecret)
			if err != nil {
				return fmt.Errorf("failed to save secret")
			}

			return nil
		},
	}

	cmd.Flags().String("secret_id", "", "ID of the text secret")
	cmd.Flags().String("content", "", "Text content of the secret")
	cmd.Flags().StringToString("meta", nil, "Optional metadata key=value pairs")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")
	_ = cmd.MarkFlagRequired("content")

	return cmd
}

func parseAddTextFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Text, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")

	secretID, _ := cmd.Flags().GetString("secret_id")
	content, _ := cmd.Flags().GetString("content")
	metaFlag, _ := cmd.Flags().GetStringToString("meta")

	if interactive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter secret_id: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		secretID = strings.TrimSpace(input)

		fmt.Print("Enter content: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		content = strings.TrimSpace(input)

		fmt.Print("Enter meta (key=value pairs separated by commas, optional): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		metaFlag = parseAddMetaString(strings.TrimSpace(input))
	}

	if secretID == "" {
		return nil, nil, fmt.Errorf("secret_id required")
	}
	if content == "" {
		return nil, nil, fmt.Errorf("content required")
	}

	config, err := configs.NewClientConfig(configs.WithDB())
	if err != nil {
		return nil, nil, fmt.Errorf("db connection error")
	}

	req := &models.Text{
		SecretID:  secretID,
		Content:   content,
		Meta:      metaFlag,
		UpdatedAt: time.Now().UTC(),
	}

	return config, req, nil
}

// newAddBinarySecretCommand creates a Cobra command to add a binary secret.
func newAddBinarySecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Add a binary secret (e.g., a file)",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, binarySecret, err := parseAddBinaryFlags(cmd)
			if err != nil {
				return fmt.Errorf("input error")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = services.AddBinary(ctx, config.DB, binarySecret)
			if err != nil {
				return fmt.Errorf("failed to save secret")
			}

			return nil
		},
	}

	cmd.Flags().String("secret_id", "", "ID of the binary secret")
	cmd.Flags().String("file", "", "Path to the binary file")
	cmd.Flags().StringToString("meta", nil, "Optional metadata key=value pairs")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")
	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func parseAddBinaryFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Binary, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")

	secretID, _ := cmd.Flags().GetString("secret_id")
	filePath, _ := cmd.Flags().GetString("file")
	metaFlag, _ := cmd.Flags().GetStringToString("meta")

	if interactive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter secret_id: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		secretID = strings.TrimSpace(input)

		fmt.Print("Enter file path: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		filePath = strings.TrimSpace(input)

		fmt.Print("Enter meta (key=value pairs separated by commas, optional): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		metaFlag = parseAddMetaString(strings.TrimSpace(input))
	}

	if secretID == "" {
		return nil, nil, fmt.Errorf("secret_id required")
	}
	if filePath == "" {
		return nil, nil, fmt.Errorf("file path required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("file read error")
	}

	config, err := configs.NewClientConfig(configs.WithDB())
	if err != nil {
		return nil, nil, fmt.Errorf("db connection error")
	}

	req := &models.Binary{
		SecretID:  secretID,
		Data:      data,
		Meta:      metaFlag,
		UpdatedAt: time.Now().UTC(),
	}

	return config, req, nil
}

// newAddCardSecretCommand creates a Cobra command to add a card secret.
func newAddCardSecretCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-card",
		Short: "Add a card secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, cardSecret, err := parseAddCardFlags(cmd)
			if err != nil {
				return fmt.Errorf("input error")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err = services.AddCard(ctx, config.DB, cardSecret)
			if err != nil {
				return fmt.Errorf("failed to save secret")
			}

			fmt.Println("Card secret added successfully")
			return nil
		},
	}

	cmd.Flags().String("secret_id", "", "ID of the card secret")
	cmd.Flags().String("number", "", "Card number")
	cmd.Flags().String("holder", "", "Cardholder name")
	cmd.Flags().Int("exp_month", 0, "Expiration month (1-12)")
	cmd.Flags().Int("exp_year", 0, "Expiration year (e.g. 2025)")
	cmd.Flags().String("cvv", "", "CVV code")
	cmd.Flags().StringToString("meta", nil, "Optional metadata")
	cmd.Flags().Bool("interactive", false, "Enable interactive input")

	_ = cmd.MarkFlagRequired("secret_id")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("holder")
	_ = cmd.MarkFlagRequired("exp_month")
	_ = cmd.MarkFlagRequired("exp_year")
	_ = cmd.MarkFlagRequired("cvv")

	return cmd
}

func parseAddCardFlags(cmd *cobra.Command) (*configs.ClientConfig, *models.Card, error) {
	interactive, _ := cmd.Flags().GetBool("interactive")

	secretID, _ := cmd.Flags().GetString("secret_id")
	number, _ := cmd.Flags().GetString("number")
	holder, _ := cmd.Flags().GetString("holder")
	expMonth, _ := cmd.Flags().GetInt("exp_month")
	expYear, _ := cmd.Flags().GetInt("exp_year")
	cvv, _ := cmd.Flags().GetString("cvv")
	metaFlag, _ := cmd.Flags().GetStringToString("meta")

	if interactive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter secret_id: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		secretID = strings.TrimSpace(input)

		fmt.Print("Enter card number: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		number = strings.TrimSpace(input)

		fmt.Print("Enter cardholder name: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		holder = strings.TrimSpace(input)

		fmt.Print("Enter expiration month (1-12): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		monthStr := strings.TrimSpace(input)
		expMonth, err = strconv.Atoi(monthStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid expiration month")
		}

		fmt.Print("Enter expiration year (e.g. 2025): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		yearStr := strings.TrimSpace(input)
		expYear, err = strconv.Atoi(yearStr)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid expiration year")
		}

		fmt.Print("Enter CVV: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		cvv = strings.TrimSpace(input)

		fmt.Print("Enter meta (key=value pairs separated by commas, optional): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("input error")
		}
		metaFlag = parseAddMetaString(strings.TrimSpace(input))
	}

	if secretID == "" {
		return nil, nil, fmt.Errorf("secret_id required")
	}
	if number == "" {
		return nil, nil, fmt.Errorf("card number required")
	}
	if holder == "" {
		return nil, nil, fmt.Errorf("cardholder required")
	}
	if expMonth < 1 || expMonth > 12 {
		return nil, nil, fmt.Errorf("invalid exp_month")
	}
	if expYear < 2000 || expYear > 2100 {
		return nil, nil, fmt.Errorf("invalid exp_year")
	}
	if cvv == "" {
		return nil, nil, fmt.Errorf("cvv required")
	}

	config, err := configs.NewClientConfig(configs.WithDB())
	if err != nil {
		return nil, nil, fmt.Errorf("db connection error")
	}

	req := &models.Card{
		SecretID:  secretID,
		Number:    number,
		Holder:    holder,
		ExpMonth:  expMonth,
		ExpYear:   expYear,
		CVV:       cvv,
		Meta:      metaFlag,
		UpdatedAt: time.Now().UTC(),
	}

	return config, req, nil
}

func parseAddMetaString(s string) map[string]string {
	m := make(map[string]string)
	if s == "" {
		return m
	}
	parts := strings.Split(s, ",")
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			m[key] = value
		}
	}
	return m
}
