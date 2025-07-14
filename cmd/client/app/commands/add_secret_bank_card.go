package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/flags"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"

	"github.com/spf13/cobra"
)

// RegisterAddSecretBankCardCommand регистрирует команду add-bank-card для добавления
// нового секрета типа банковская карта. Команда принимает необходимые параметры,
// создает модель секрета и сохраняет её в локальное хранилище.
//
// Параметры команды:
//
//	--secret-name  (обязательный) имя секрета
//	--number       (обязательный) номер карты
//	--owner        (необязательный) владелец карты
//	--exp          (обязательный) срок действия карты в формате MM/YY
//	--cvv          (обязательный) CVV-код карты
//	--meta         (необязательный) дополнительная метаинформация в формате JSON
func RegisterAddSecretBankCardCommand(root *cobra.Command) {
	var secretName string
	var number string
	var owner string
	var exp string
	var cvv string
	var meta string

	cmd := &cobra.Command{
		Use:   "add-secret-bank-card",
		Short: "Add secret: bank card",
		RunE: func(cmd *cobra.Command, args []string) error {
			if secretName == "" {
				return errors.New("параметр --secret-name обязателен")
			}
			if number == "" {
				return errors.New("параметр --number обязателен")
			}
			if exp == "" {
				return errors.New("параметр --exp обязателен")
			}
			if cvv == "" {
				return errors.New("параметр --cvv обязателен")
			}

			config, err := configs.NewClientConfig(
				configs.WithDB("gophkeeper.db"),
			)
			if err != nil {
				return err
			}

			metaJSON, err := flags.PrepareMetaJSON(meta)
			if err != nil {
				return fmt.Errorf("failed to parse meta JSON")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			repo := repositories.NewSecretBankCardClientSaveRepository(config.DB)

			secret := models.SecretBankCardClient{
				SecretName: secretName,
				Number:     number,
				Owner:      owner,
				Exp:        exp,
				CVV:        cvv,
				Meta:       metaJSON,
				UpdatedAt:  time.Now(),
			}

			return repo.Save(ctx, secret)
		},
	}

	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name (required)")
	cmd.Flags().StringVar(&number, "number", "", "Card number (required)")
	cmd.Flags().StringVar(&owner, "owner", "", "Card owner (optional)")
	cmd.Flags().StringVar(&exp, "exp", "", "Card expiration date MM/YY (required)")
	cmd.Flags().StringVar(&cvv, "cvv", "", "Card CVV code (required)")
	cmd.Flags().StringVar(&meta, "meta", "", "Additional meta JSON (optional)")

	_ = cmd.MarkFlagRequired("secret-name")
	_ = cmd.MarkFlagRequired("number")
	_ = cmd.MarkFlagRequired("exp")
	_ = cmd.MarkFlagRequired("cvv")

	root.AddCommand(cmd)
}
