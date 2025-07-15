package commands

import (
	"context"
	"errors"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"

	"github.com/spf13/cobra"
)

func RegisterAddSecretTextCommand(root *cobra.Command) {
	var secretName string
	var content string
	var meta string

	cmd := &cobra.Command{
		Use:   "add-secret-text",
		Short: "Добавить секрет: текстовый",
		RunE: func(cmd *cobra.Command, args []string) error {
			if secretName == "" {
				return errors.New("ошибка: параметр --secret-name обязателен")
			}
			if content == "" {
				return errors.New("ошибка: параметр --content обязателен")
			}

			cfg, err := configs.NewClientConfig(
				configs.WithClientConfigDB("gophkeeper.db"),
			)
			if err != nil {
				return errors.New("ошибка: не удалось инициализировать конфигурацию клиента")
			}

			metaJSON, err := configs.PrepareMetaJSON(meta)
			if err != nil {
				return errors.New("ошибка: неверный формат параметра --meta")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			secret := models.SecretTextSaveRequest{
				SecretName: secretName,
				Content:    content,
				Meta:       metaJSON,
			}

			err = client.SaveSecretTextRequest(ctx, cfg.DB, secret)
			if err != nil {
				return errors.New("ошибка: не удалось сохранить данные в базу")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Имя секрета (обязательно)")
	cmd.Flags().StringVar(&content, "content", "", "Содержимое секрета (обязательно)")
	cmd.Flags().StringVar(&meta, "meta", "", "Доп. метаданные JSON (необязательно)")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("content")

	root.AddCommand(cmd)
}
