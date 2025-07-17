package commands

import (
	"context"

	"github.com/spf13/cobra"
)

func RegisterAddTextSecretCommand(root *cobra.Command) {
	// Локальные переменные для флагов
	var (
		secretName string
		content    string
		meta       string

		authURL string
		tlsCert string
		token   string
	)

	cmd := &cobra.Command{
		Use:   "add-text-secret",
		Short: "Add a text secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := runAddText(cmd.Context(), secretName, content, meta, authURL, tlsCert, token)
			if err != nil {
				return err
			}
			cmd.Println(res)
			return nil
		},
	}

	// Регистрируем флаги
	cmd.Flags().StringVar(&secretName, "secret-name", "", "Secret name")
	cmd.Flags().StringVar(&content, "content", "", "Text content")
	cmd.Flags().StringVar(&meta, "meta", "", "Optional metadata")

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	root.AddCommand(cmd)
}

// addTextSecret принимает параметры явно
func runAddText(
	ctx context.Context,
	secretName string,
	content string,
	meta string,
	authURL string,
	tlsCert string,
	token string,
) (string, error) {
	// TODO: реализовать логику добавления текстового секрета
	return "Text secret added successfully", nil
}
