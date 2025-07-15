package commands

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"

	"github.com/spf13/cobra"
)

func RegisterAddSecretBinaryCommand(root *cobra.Command) {
	var secretName string
	var dataPath string
	var meta string

	cmd := &cobra.Command{
		Use:   "add-secret-binary",
		Short: "Добавить секрет: бинарные данные",
		RunE: func(cmd *cobra.Command, args []string) error {
			if secretName == "" {
				return errors.New("ошибка: параметр --secret-name обязателен")
			}
			if dataPath == "" {
				return errors.New("ошибка: параметр --data обязателен")
			}

			cfg, err := configs.NewClientConfig(
				configs.WithClientConfigDB("gophkeeper.db"),
			)
			if err != nil {
				return errors.New("ошибка: не удалось инициализировать конфигурацию клиента")
			}

			data, err := os.ReadFile(dataPath)
			if err != nil {
				return errors.New("ошибка: не удалось прочитать файл с данными")
			}

			metaJSON, err := configs.PrepareMetaJSON(meta)
			if err != nil {
				return errors.New("ошибка: неверный формат параметра --meta")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			secret := models.SecretBinarySaveRequest{
				SecretName: secretName,
				Data:       data,
				Meta:       metaJSON,
			}

			err = client.SaveSecretBinaryRequest(ctx, cfg.DB, secret)
			if err != nil {
				return errors.New("ошибка: не удалось сохранить данные в базу")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Имя секрета (обязательно)")
	cmd.Flags().StringVar(&dataPath, "data", "", "Путь к файлу с данными (обязательно)")
	cmd.Flags().StringVar(&meta, "meta", "", "Доп. метаданные JSON (необязательно)")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("data")

	root.AddCommand(cmd)
}
