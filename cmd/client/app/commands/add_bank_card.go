package commands

import (
	"context"

	"github.com/spf13/cobra"
)

func RegisterAddBankCardSecretCommand(root *cobra.Command) {
	// Локальные переменные для флагов
	var (
		secretName string
		number     string
		owner      string
		exp        string
		cvv        string
		meta       string

		authURL string
		tlsCert string
		token   string
	)

	cmd := &cobra.Command{
		Use:   "add-bank-card",
		Short: "Add a bank card secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := runAddBankCard(cmd.Context(), secretName, number, owner, exp, cvv, meta, authURL, tlsCert, token)
			if err != nil {
				return err
			}
			cmd.Println(res)
			return nil
		},
	}

	// Регистрируем флаги и связываем с локальными переменными
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&number, "number", "", "Bank card number")
	cmd.Flags().StringVar(&owner, "owner", "", "Bank card owner")
	cmd.Flags().StringVar(&exp, "exp", "", "Bank card expiration date")
	cmd.Flags().StringVar(&cvv, "cvv", "", "Bank card CVV code")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	root.AddCommand(cmd)
}

// addBankCard принимает все параметры явно
func runAddBankCard(
	ctx context.Context,
	secretNamestring,
	numberstring,
	ownerstring,
	expstring,
	cvv string,
	meta string,
	authURL string,
	tlsCert string,
	token string,
) (string, error) {
	// TODO: реализовать логику с использованием этих параметров
	return "Bank card added successfully", nil
}
