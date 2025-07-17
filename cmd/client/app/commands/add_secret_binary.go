package commands

import (
	"context"

	"github.com/spf13/cobra"
)

func RegisterAddBinarySecretCommand(root *cobra.Command) {
	// Локальные переменные для флагов
	var (
		secretName string
		dataPath   string
		meta       string

		authURL string
		tlsCert string
		token   string
	)

	cmd := &cobra.Command{
		Use:   "add-binary-secret",
		Short: "Add a binary secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := runAddBinary(cmd.Context(), secretName, dataPath, meta, authURL, tlsCert, token)
			if err != nil {
				return err
			}
			cmd.Println(res)
			return nil
		},
	}

	// Регистрируем флаги и связываем с локальными переменными
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&dataPath, "data-path", "", "Path to binary data file")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	root.AddCommand(cmd)
}

// addBinarySecret принимает параметры явно
func runAddBinary(
	ctx context.Context,
	secretName string,
	dataPath string,
	meta string,
	authURL string,
	tlsCert string,
	token string,
) (string, error) {
	// TODO: реализовать логику добавления бинарного секрета
	return "Binary secret added successfully", nil
}
