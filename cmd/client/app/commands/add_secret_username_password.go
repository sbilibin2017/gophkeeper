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

// RegisterAddSecretUsernamePasswordCommand регистрирует команду add-username-password,
// которая добавляет секрет с логином и паролем.
//
// Параметры команды:
//
//	--secret-name (обязательный) — название секрета
//	--username    (обязательный) — имя пользователя
//	--password    (обязательный) — пароль
//	--meta        (необязательный) — дополнительные данные в формате JSON
func RegisterAddSecretUsernamePasswordCommand(root *cobra.Command) {
	var secretName string
	var username string
	var password string
	var meta string

	cmd := &cobra.Command{
		Use:   "add-secret-username-password",
		Short: "Добавить секрет: логин и пароль",
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

			repo := repositories.NewSecretUsernamePasswordClientSaveRepository(config.DB)

			secret := models.SecretUsernamePasswordClient{
				SecretName: secretName,
				Username:   username,
				Password:   password,
				Meta:       metaJSON,
				UpdatedAt:  time.Now(),
			}

			return repo.Save(ctx, secret)
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Название секрета (обязательный параметр)")
	cmd.Flags().StringVar(&username, "username", "", "Имя пользователя (обязательный параметр)")
	cmd.Flags().StringVar(&password, "password", "", "Пароль (обязательный параметр)")
	cmd.Flags().StringVar(&meta, "meta", "", "Дополнительные данные в формате JSON (необязательно)")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")

	root.AddCommand(cmd)
}
