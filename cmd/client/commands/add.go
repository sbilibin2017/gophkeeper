package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"

	"github.com/spf13/cobra"
)

func NewAddBankCardCommand() *cobra.Command {
	var (
		certPath   string
		secretName string
		number     string
		owner      string
		exp        string
		cvv        string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-bankcard",
		Short: "Adds bankcard secret to client",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			db, err := db.New("sqlite", "client.db",
				db.WithMaxOpenConns(10),
				db.WithMaxIdleConns(5),
				db.WithConnMaxLifetime(30*time.Minute),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer db.Close()

			writeRepo := repositories.NewEncryptedSecretWriteRepository(db)

			cryptor, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(certPath),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize cryptor: %w", err)
			}

			secretWriter := client.NewSecretWriter(writeRepo, cryptor)

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			bankCardPayload := models.BankCardPayload{
				Number: number,
				Owner:  owner,
				Exp:    exp,
				CVV:    cvv,
				Meta:   metaPtr,
			}

			if err := secretWriter.AddBankCard(ctx, secretName, bankCardPayload); err != nil {
				return fmt.Errorf("failed to add bank card secret: %w", err)
			}

			fmt.Println("Bank card secret added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&certPath, "cert", "", "Path to RSA public certificate PEM file (required)")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Name of the secret (required)")
	cmd.Flags().StringVar(&number, "number", "", "Bank card number (required)")
	cmd.Flags().StringVar(&owner, "owner", "", "Bank card owner name (required)")
	cmd.Flags().StringVar(&exp, "exp", "", "Bank card expiration date (MM/YY) (required)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "Bank card CVV (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Additional metadata (optional)")

	cmd.MarkFlagRequired("cert")
	cmd.MarkFlagRequired("secret-name")
	cmd.MarkFlagRequired("number")
	cmd.MarkFlagRequired("owner")
	cmd.MarkFlagRequired("exp")
	cmd.MarkFlagRequired("cvv")

	return cmd
}

func NewAddBinaryCommand() *cobra.Command {
	var (
		certPath   string
		secretName string
		filePath   string
		dataPath   string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Adds binary secret to client",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			data, err := os.ReadFile(dataPath)
			if err != nil {
				return fmt.Errorf("failed to read binary data file: %w", err)
			}

			db, err := db.New("sqlite", "client.db",
				db.WithMaxOpenConns(10),
				db.WithMaxIdleConns(5),
				db.WithConnMaxLifetime(30*time.Minute),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer db.Close()

			writeRepo := repositories.NewEncryptedSecretWriteRepository(db)

			cryptor, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(certPath),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize cryptor: %w", err)
			}

			secretWriter := client.NewSecretWriter(writeRepo, cryptor)

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			binaryPayload := models.BinaryPayload{
				FileName: filePath,
				Data:     data,
				Meta:     metaPtr,
			}

			if err := secretWriter.AddBinary(ctx, secretName, binaryPayload); err != nil {
				return fmt.Errorf("failed to add binary secret: %w", err)
			}

			fmt.Println("Binary secret added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&certPath, "cert", "", "Path to RSA public certificate PEM file (required)")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name (required)")
	cmd.Flags().StringVar(&filePath, "file-path", "", "File name for the binary secret (required)")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "Path to binary data file (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Additional metadata (optional)")

	cmd.MarkFlagRequired("cert")
	cmd.MarkFlagRequired("secret-name")
	cmd.MarkFlagRequired("file-path")
	cmd.MarkFlagRequired("data-path")

	return cmd
}

func NewAddTextCommand() *cobra.Command {
	var (
		certPath   string
		secretName string
		data       string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-text",
		Short: "Adds text secret to client",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			db, err := db.New("sqlite", "client.db",
				db.WithMaxOpenConns(10),
				db.WithMaxIdleConns(5),
				db.WithConnMaxLifetime(30*time.Minute),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer db.Close()

			writeRepo := repositories.NewEncryptedSecretWriteRepository(db)

			cryptor, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(certPath),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize cryptor: %w", err)
			}

			secretWriter := client.NewSecretWriter(writeRepo, cryptor)

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			textPayload := models.TextPayload{
				Data: data,
				Meta: metaPtr,
			}

			if err := secretWriter.AddText(ctx, secretName, textPayload); err != nil {
				return fmt.Errorf("failed to add text secret: %w", err)
			}

			fmt.Println("Text secret added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&certPath, "cert", "", "Path to RSA public certificate PEM file (required)")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name (required)")
	cmd.Flags().StringVar(&data, "data", "", "Text data (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Additional metadata (optional)")

	cmd.MarkFlagRequired("cert")
	cmd.MarkFlagRequired("secret-name")
	cmd.MarkFlagRequired("data")

	return cmd
}

func NewAddUserCommand() *cobra.Command {
	var (
		certPath   string
		secretName string
		login      string
		password   string
		meta       string
	)

	cmd := &cobra.Command{
		Use:   "add-user",
		Short: "Adds user secret to client",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			db, err := db.New("sqlite", "client.db",
				db.WithMaxOpenConns(10),
				db.WithMaxIdleConns(5),
				db.WithConnMaxLifetime(30*time.Minute),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to DB: %w", err)
			}
			defer db.Close()

			writeRepo := repositories.NewEncryptedSecretWriteRepository(db)

			cryptor, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(certPath),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize cryptor: %w", err)
			}

			secretWriter := client.NewSecretWriter(writeRepo, cryptor)

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			userPayload := models.UserPayload{
				Login:    login,
				Password: password,
				Meta:     metaPtr,
			}

			if err := secretWriter.AddUser(ctx, secretName, userPayload); err != nil {
				return fmt.Errorf("failed to add user secret: %w", err)
			}

			fmt.Println("User secret added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&certPath, "cert", "", "Path to RSA public certificate PEM file (required)")
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name (required)")
	cmd.Flags().StringVar(&login, "login", "", "Login (required)")
	cmd.Flags().StringVar(&password, "password", "", "Password (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Additional metadata (optional)")

	cmd.MarkFlagRequired("cert")
	cmd.MarkFlagRequired("secret-name")
	cmd.MarkFlagRequired("login")
	cmd.MarkFlagRequired("password")

	return cmd
}
