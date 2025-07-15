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

func RegisterAddSecretUsernamePasswordCommand(root *cobra.Command) {
	var secretName string
	var username string
	var password string
	var meta string

	cmd := &cobra.Command{
		Use:   "add-secret-username-password",
		Short: "Добавить секрет: логин и пароль",
		RunE: func(cmd *cobra.Command, args []string) error {
			if secretName == "" {
				return errors.New("ошибка: параметр --secret-name обязателен")
			}
			if username == "" {
				return errors.New("ошибка: параметр --username обязателен")
			}
			if password == "" {
				return errors.New("ошибка: параметр --password обязателен")
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

			secret := models.SecretUsernamePasswordSaveRequest{
				SecretName: secretName,
				Username:   username,
				Password:   password,
				Meta:       metaJSON,
			}

			err = client.SaveSecretUsernamePasswordRequest(ctx, cfg.DB, secret)
			if err != nil {
				return errors.New("ошибка: не удалось сохранить данные в базу")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Имя секрета (обязательно)")
	cmd.Flags().StringVar(&username, "username", "", "Имя пользователя (обязательно)")
	cmd.Flags().StringVar(&password, "password", "", "Пароль (обязательно)")
	cmd.Flags().StringVar(&meta, "meta", "", "Доп. метаданные JSON (необязательно)")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	root.AddCommand(cmd)
}
