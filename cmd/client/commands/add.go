package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/cryptor"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"

	"github.com/spf13/cobra"
)

// NewAddBankCardCommand creates a new Cobra command to add a bank card secret to the local client storage.
// It accepts various flags like secret name, card number, owner, expiration, and CVV.
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

			writer := repositories.NewEncryptedSecretWriteRepository(db)

			cryptor, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(certPath),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize cryptor: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			payload := models.BankCardPayload{
				Number: number,
				Owner:  owner,
				Exp:    exp,
				CVV:    cvv,
				Meta:   metaPtr,
			}

			plaintext, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to marshal bank card payload: %w", err)
			}

			enc, err := cryptor.Encrypt(plaintext)
			if err != nil {
				return fmt.Errorf("encryption failed: %w", err)
			}

			secret := &models.EncryptedSecret{
				SecretType: models.SecretTypeBankCard,
				SecretName: secretName,
				Ciphertext: enc.Ciphertext,
				AESKeyEnc:  enc.AESKeyEnc,
			}

			if err := writer.Save(ctx, secret); err != nil {
				return fmt.Errorf("failed to save bank card secret: %w", err)
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

// NewAddBinaryCommand creates a Cobra command to add binary file data as a secret.
// Requires input of file name, data path, and encryption certificate. Optionally accepts metadata.
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

			writer := repositories.NewEncryptedSecretWriteRepository(db)

			cryptor, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(certPath),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize cryptor: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			payload := models.BinaryPayload{
				FileName: filePath,
				Data:     data,
				Meta:     metaPtr,
			}

			plaintext, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to marshal binary payload: %w", err)
			}

			enc, err := cryptor.Encrypt(plaintext)
			if err != nil {
				return fmt.Errorf("encryption failed: %w", err)
			}

			secret := &models.EncryptedSecret{
				SecretType: models.SecretTypeBinary,
				SecretName: secretName,
				Ciphertext: enc.Ciphertext,
				AESKeyEnc:  enc.AESKeyEnc,
			}

			if err := writer.Save(ctx, secret); err != nil {
				return fmt.Errorf("failed to save binary secret: %w", err)
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

// NewAddTextCommand creates a Cobra command to store plain text secrets in the client database.
// Requires secret name and text content. Supports optional metadata.
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

			writer := repositories.NewEncryptedSecretWriteRepository(db)

			cryptor, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(certPath),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize cryptor: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			payload := models.TextPayload{
				Data: data,
				Meta: metaPtr,
			}

			plaintext, err := json.Marshal(payload)
			if err != nil {
				return fmt.Errorf("failed to marshal text payload: %w", err)
			}

			enc, err := cryptor.Encrypt(plaintext)
			if err != nil {
				return fmt.Errorf("encryption failed: %w", err)
			}

			secret := &models.EncryptedSecret{
				SecretType: models.SecretTypeText,
				SecretName: secretName,
				Ciphertext: enc.Ciphertext,
				AESKeyEnc:  enc.AESKeyEnc,
			}

			if err := writer.Save(ctx, secret); err != nil {
				return fmt.Errorf("failed to save text secret: %w", err)
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

// NewAddUserCommand creates a Cobra command to add user credentials (login and password) as a secret.
// Requires RSA public key, secret name, login, and password. Metadata is optional.
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

			writer := repositories.NewEncryptedSecretWriteRepository(db)

			cryptor, err := cryptor.New(
				cryptor.WithPublicKeyFromCert(certPath),
			)
			if err != nil {
				return fmt.Errorf("failed to initialize cryptor: %w", err)
			}

			var metaPtr *string
			if meta != "" {
				metaPtr = &meta
			}

			payload := models.UserPayload{
				Login:    login,
				Password: password,
				Meta:     metaPtr,
			}

			plaintext, err := json.Marshal(payload)
			if err != nil {
				return err
			}

			enc, err := cryptor.Encrypt(plaintext)
			if err != nil {
				return err
			}

			secret := &models.EncryptedSecret{
				SecretType: models.SecretTypeUser,
				SecretName: secretName,
				Ciphertext: enc.Ciphertext,
				AESKeyEnc:  enc.AESKeyEnc,
			}

			if err := writer.Save(ctx, secret); err != nil {
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
