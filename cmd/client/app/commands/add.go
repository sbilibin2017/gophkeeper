package commands

import (
	"os"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

// RegisterAddBankCardCommand registers the 'add-bank-card' command to the root command.
// This command allows users to add a bank card secret with fields like number, owner, expiration, CVV, and optional metadata.
func RegisterAddBankCardCommand(root *cobra.Command) {
	var (
		secretName string
		number     string
		owner      string
		exp        string
		cvv        string
		meta       string
		token      string
	)

	cmd := &cobra.Command{
		Use:   "add-bank-card",
		Short: "Add a bank card secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := configs.NewClientConfig(configs.WithClientConfigDB())
			if err != nil {
				return err
			}

			req := models.BankCardAddRequest{
				SecretName: secretName,
				Number:     number,
				Owner:      owner,
				Exp:        exp,
				CVV:        cvv,
			}
			if meta != "" {
				req.Meta = &meta
			}

			ctx := cmd.Context()

			err = client.AddBankCardSecret(ctx, config.DB, req)
			if err != nil {
				return err
			}

			cmd.Println("Bank card secret added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&number, "number", "", "Bank card number")
	cmd.Flags().StringVar(&owner, "owner", "", "Bank card owner")
	cmd.Flags().StringVar(&exp, "exp", "", "Bank card expiration date")
	cmd.Flags().StringVar(&cvv, "cvv", "", "Bank card CVV code")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("exp")
	_ = cmd.MarkFlagRequired("cvv")
	_ = cmd.MarkFlagRequired("token")

	root.AddCommand(cmd)
}

// RegisterAddBinaryCommand registers the 'add-binary-secret' command to the root command.
// This command lets users add a binary secret by specifying a file path and optional metadata.
func RegisterAddBinaryCommand(root *cobra.Command) {
	var (
		secretName string
		dataPath   string
		meta       string
		token      string
	)

	cmd := &cobra.Command{
		Use:   "add-binary-secret",
		Short: "Add a binary secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := configs.NewClientConfig(configs.WithClientConfigDB())
			if err != nil {
				return err
			}

			data, err := os.ReadFile(dataPath)
			if err != nil {
				return err
			}

			req := models.AddSecretBinaryRequest{
				SecretName: secretName,
				Data:       data,
			}
			if meta != "" {
				req.Meta = &meta
			}

			ctx := cmd.Context()

			err = client.AddBinarySecret(ctx, config.DB, req)
			if err != nil {
				return err
			}

			cmd.Println("Binary secret added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "Path to binary data file")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("data-path")
	_ = cmd.MarkFlagRequired("token")

	root.AddCommand(cmd)
}

// RegisterAddTextCommand registers the 'add-text-secret' command to the root command.
// This command allows users to add a text secret with content and optional metadata.
func RegisterAddTextCommand(root *cobra.Command) {
	var (
		secretName string
		content    string
		meta       string
		token      string
	)

	cmd := &cobra.Command{
		Use:   "add-text-secret",
		Short: "Add a text secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := configs.NewClientConfig(configs.WithClientConfigDB())
			if err != nil {
				return err
			}

			req := models.TextAddRequest{
				SecretName: secretName,
				Content:    content,
			}
			if meta != "" {
				req.Meta = &meta
			}

			ctx := cmd.Context()

			err = client.AddTextSecret(ctx, config.DB, req)
			if err != nil {
				return err
			}

			cmd.Println("Text secret added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&content, "content", "", "Text content")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("content")
	_ = cmd.MarkFlagRequired("token")

	root.AddCommand(cmd)
}

// RegisterAddUsernamePasswordCommand registers the 'add-username-password' command to the root command.
// This command enables users to add a username-password secret with optional metadata.
func RegisterAddUsernamePasswordCommand(root *cobra.Command) {
	var (
		secretName string
		user       string
		pass       string
		meta       string
		token      string
	)

	cmd := &cobra.Command{
		Use:   "add-username-password",
		Short: "Add a username-password secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := configs.NewClientConfig(configs.WithClientConfigDB())
			if err != nil {
				return err
			}

			req := models.UsernamePasswordAddRequest{
				SecretName: secretName,
				Username:   user,
				Password:   pass,
			}
			if meta != "" {
				req.Meta = &meta
			}

			ctx := cmd.Context()

			err = client.AddUsernamePasswordSecret(ctx, config.DB, req)
			if err != nil {
				return err
			}

			cmd.Println("Username-password secret added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&user, "user", "", "Username for username-password secret")
	cmd.Flags().StringVar(&pass, "pass", "", "Password for username-password secret")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("user")
	_ = cmd.MarkFlagRequired("pass")
	_ = cmd.MarkFlagRequired("token")

	root.AddCommand(cmd)
}
