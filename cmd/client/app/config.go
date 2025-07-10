package app

import (
	"fmt"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/config"
	"github.com/spf13/cobra"
)

// newConfigCommand создаёт команду для настройки параметров клиента.
func newConfigCommand() *cobra.Command {
	var token string
	var serverURL string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Настройка параметров клиента",
		Long: `Команда для настройки параметров клиента Gophkeeper.

Позволяет сохранить токен аутентификации и/или URL сервера в
соответствующих переменных окружения или конфигурационных файлах.

Использование:

  gophkeeper config --token mytoken123
  gophkeeper config --server-url https://example.com
  gophkeeper config --token mytoken123 --server-url https://example.com
`,
		Example: `  gophkeeper config --token mytoken123
  gophkeeper config --server-url https://example.com
  gophkeeper config --token mytoken123 --server-url https://example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if token == "" && serverURL == "" {
				return fmt.Errorf("необходимо указать хотя бы один флаг: --token или --server-url")
			}

			if token != "" {
				if err := config.SetToken(token); err != nil {
					return err
				}
				fmt.Println("Токен сохранён в GOPHKEEPER_TOKEN")
			}

			if serverURL != "" {
				if err := config.SetServerURL(serverURL); err != nil {
					return err
				}
				fmt.Println("URL сервера сохранён в GOPHKEEPER_SERVER_URL")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&token, "token", "", "Токен аутентификации")
	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL сервера")

	return cmd
}
