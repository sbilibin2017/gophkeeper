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

func RegisterAddSecretBankCardCommand(root *cobra.Command) {
	var secretName string
	var number string
	var owner string
	var exp string
	var cvv string
	var meta string

	cmd := &cobra.Command{
		Use:   "add-secret-bank-card",
		Short: "Добавить секрет: банковская карта",
		RunE: func(cmd *cobra.Command, args []string) error {
			if secretName == "" {
				return errors.New("ошибка: параметр --secret-name обязателен")
			}
			if number == "" {
				return errors.New("ошибка: параметр --number обязателен")
			}
			if exp == "" {
				return errors.New("ошибка: параметр --exp обязателен")
			}
			if cvv == "" {
				return errors.New("ошибка: параметр --cvv обязателен")
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

			secret := models.SecretBankCardSaveRequest{
				SecretName: secretName,
				Number:     number,
				Owner:      owner,
				Exp:        exp,
				CVV:        cvv,
				Meta:       metaJSON,
			}

			err = client.SaveSecretBankCardRequest(ctx, cfg.DB, secret)
			if err != nil {
				return errors.New("ошибка: не удалось сохранить данные в базу")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Имя секрета (обязательно)")
	cmd.Flags().StringVar(&number, "number", "", "Номер карты (обязательно)")
	cmd.Flags().StringVar(&owner, "owner", "", "Владелец карты (необязательно)")
	cmd.Flags().StringVar(&exp, "exp", "", "Срок действия MM/YY (обязательно)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "CVV код (обязательно)")
	cmd.Flags().StringVar(&meta, "meta", "", "Доп. метаданные JSON (необязательно)")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("exp")
	_ = cmd.MarkFlagRequired("cvv")

	root.AddCommand(cmd)
}
