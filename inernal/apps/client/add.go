package client

import (
	"context"
	"fmt"
	"os"

	"github.com/sbilibin2017/gophkeeper/inernal/models"
	"github.com/sbilibin2017/gophkeeper/inernal/usecases/client"
	"github.com/spf13/cobra"
)

func NewAddBankCardCommand(uc *client.SecretClientAddUsecase, rsaPubKeyPath string) *cobra.Command {
	var (
		token      string
		secretName string
		number     string
		owner      string
		exp        string
		cvv        string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-bankcard",
		Short: "Add a new bank card entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			secret := &models.BankcardSecretAdd{
				SecretName:  secretName,
				SecretType:  models.SecretTypeBankCard,
				SecretOwner: token,
				Number:      number,
				Owner:       owner,
				Exp:         exp,
				CVV:         cvv,
				Meta:        metaPtr,
			}

			if err := uc.AddBankCard(ctx, secret, token); err != nil {
				return fmt.Errorf("failed to save secret: %w", err)
			}

			fmt.Println("Bank card secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "JWT token for authentication (required)")
	cmd.Flags().StringVar(&secretName, "name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&number, "number", "", "Card number (required)")
	cmd.Flags().StringVar(&owner, "owner", "", "Card owner's name (required)")
	cmd.Flags().StringVar(&exp, "exp", "", "Card expiration date (MM/YY) (required)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "Card CVV code (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("exp")
	_ = cmd.MarkFlagRequired("cvv")

	return cmd
}

func NewAddBinaryCommand(uc *client.SecretClientAddUsecase) *cobra.Command {
	var (
		token      string
		secretName string
		dataPath   string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Add a new binary data entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if secretName == "" || token == "" || dataPath == "" {
				return fmt.Errorf("flags 'name', 'token', and 'data-path' are required")
			}

			data, err := os.ReadFile(dataPath)
			if err != nil {
				return fmt.Errorf("failed to read binary data file: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			secret := &models.BinarySecretAdd{
				SecretName:  secretName,
				SecretType:  models.SecretTypeBinary,
				SecretOwner: token,
				Data:        data,
				Meta:        metaPtr,
			}

			if err := uc.AddBinarySecret(ctx, secret, token); err != nil {
				return fmt.Errorf("failed to save binary secret: %w", err)
			}

			fmt.Println("Binary secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "JWT token for authentication (required)")
	cmd.Flags().StringVar(&secretName, "name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "Path to binary data file (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("data-path")

	return cmd
}

func NewAddTextCommand(uc *client.SecretClientAddUsecase) *cobra.Command {
	var (
		token      string
		secretName string
		text       string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-text",
		Short: "Add a new text entry",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if secretName == "" || token == "" || text == "" {
				return fmt.Errorf("flags 'name', 'token', and 'text' are required")
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			secret := &models.TextSecretAdd{
				SecretName:  secretName,
				SecretType:  models.SecretTypeText,
				SecretOwner: token,
				Text:        text,
				Meta:        metaPtr,
			}

			if err := uc.AddTextSecret(ctx, secret, token); err != nil {
				return fmt.Errorf("failed to save text secret: %w", err)
			}

			fmt.Println("Text secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "JWT token for authentication (required)")
	cmd.Flags().StringVar(&secretName, "name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&text, "text", "", "Text content (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("text")

	return cmd
}

func NewAddUserCommand(uc *client.SecretClientAddUsecase) *cobra.Command {
	var (
		token      string
		secretName string
		username   string
		password   string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-user",
		Short: "Add a new user secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if secretName == "" || token == "" || username == "" || password == "" {
				return fmt.Errorf("flags 'name', 'token', 'username', and 'password' are required")
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			secret := &models.UserSecretAdd{
				SecretName:  secretName,
				SecretType:  models.SecretTypeUser,
				SecretOwner: token,
				Username:    username,
				Password:    password,
				Meta:        metaPtr,
			}

			if err := uc.AddUserSecret(ctx, secret, token); err != nil {
				return fmt.Errorf("failed to save user secret: %w", err)
			}

			fmt.Println("User secret saved successfully.")
			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "JWT token for authentication (required)")
	cmd.Flags().StringVar(&secretName, "name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&username, "username", "", "Username (required)")
	cmd.Flags().StringVar(&password, "password", "", "Password (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	return cmd
}
