package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/flags"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/spf13/cobra"
)

// RegisterAddSecretTextCommand регистрирует команду add-text-secret для добавления
// текстового секрета пользователя.
//
// Параметры команды:
//
//	--secret-name (обязательный) название секрета
//	--content     (обязательный) содержимое секрета
//	--meta        (необязательный) дополнительные данные в формате JSON
func RegisterAddSecretTextCommand(root *cobra.Command) {
	var secretName string
	var content string
	var meta string

	cmd := &cobra.Command{
		Use:   "add-secret-text",
		Short: "Добавить секрет: текстовый",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := configs.NewClientConfig(
				configs.WithDB("gophkeeper.db"),
			)
			if err != nil {
				return err
			}

			metaJSON, err := flags.PrepareMetaJSON(meta)
			if err != nil {
				return fmt.Errorf("не удалось распарсить meta")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			repo := repositories.NewSecretTextClientSaveRepository(config.DB)

			secret := models.SecretTextClient{
				SecretName: secretName,
				Content:    content,
				Meta:       metaJSON,
				UpdatedAt:  time.Now(),
			}

			return repo.Save(ctx, secret)
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Название секрета (обязательный параметр)")
	cmd.Flags().StringVar(&content, "content", "", "Содержимое секрета (обязательный параметр)")
	cmd.Flags().StringVar(&meta, "meta", "", "Дополнительные данные в формате JSON (необязательно)")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("content")

	root.AddCommand(cmd)
}
