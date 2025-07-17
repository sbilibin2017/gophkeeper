package commands

import (
	"context"

	"github.com/spf13/cobra"
)

func RegisterListSecretsCommand(root *cobra.Command) {
	var (
		secretsType string

		authURL       string
		tlsClientCert string
		tlsClientKey  string
		token         string
	)

	cmd := &cobra.Command{
		Use:   "list-secrets",
		Short: "List names of secrets of a specified type",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := runList(cmd.Context(), secretsType, authURL, tlsClientCert, tlsClientKey, token)
			if err != nil {
				return err
			}
			for _, name := range names {
				cmd.Println(name)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&secretsType, "secrets-type", "", "Type of secrets to list (bankcard, binary, text, userpass)")

	cmd.Flags().StringVar(&authURL, "auth-url", "", "Authentication service URL")
	cmd.Flags().StringVar(&tlsClientCert, "tls-client-cert", "", "Path to client TLS certificate file (optional)")
	cmd.Flags().StringVar(&tlsClientKey, "tls-client-key", "", "Path to client TLS private key file (optional)")

	cmd.Flags().StringVar(&token, "token", "", "Authentication token")

	root.AddCommand(cmd)
}

// listSecrets возвращает список имён секретов заданного типа,
// параметры принимает явно
func runList(
	ctx context.Context,
	secretsType, authURL, tlsClientCert, tlsClientKey, token string,
) ([]string, error) {
	// TODO: Реализовать логику получения списка секретов по параметрам
	return []string{}, nil
}
