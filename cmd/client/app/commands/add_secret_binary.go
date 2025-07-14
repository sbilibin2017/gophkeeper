package commands

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/flags"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"

	"github.com/spf13/cobra"
)

// RegisterAddSecretBinaryCommand регистрирует команду add-binary-secret для добавления
// нового бинарного секрета из файла. Команда принимает имя секрета, путь к файлу с данными,
// и опциональную метаинформацию в JSON формате.
//
// Параметры команды:
//
//	--secret-name (обязательный) имя секрета
//	--data        (обязательный) путь к файлу с бинарными данными
//	--meta        (необязательный) дополнительная метаинформация в формате JSON
func RegisterAddSecretBinaryCommand(root *cobra.Command) {
	var secretName string
	var dataPath string
	var meta string

	cmd := &cobra.Command{
		Use:   "add-secret-binary",
		Short: "Add secret: binary data",
		RunE: func(cmd *cobra.Command, args []string) error {
			if secretName == "" {
				return errors.New("--secret-name is required")
			}
			if dataPath == "" {
				return errors.New("--data is required")
			}

			config, err := configs.NewClientConfig(
				configs.WithDB("gophkeeper.db"),
			)
			if err != nil {
				return err
			}

			data, err := os.ReadFile(dataPath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			metaJSON, err := flags.PrepareMetaJSON(meta)
			if err != nil {
				return fmt.Errorf("failed to parse meta JSON: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			repo := repositories.NewSecretBinaryClientSaveRepository(config.DB)

			secret := models.SecretBinaryClient{
				SecretName: secretName,
				Data:       data,
				Meta:       metaJSON,
				UpdatedAt:  time.Now(),
			}

			return repo.Save(ctx, secret)
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name (required)")
	cmd.Flags().StringVar(&dataPath, "data", "", "Path to binary data file (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Additional meta JSON (optional)")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("data")

	root.AddCommand(cmd)
}
