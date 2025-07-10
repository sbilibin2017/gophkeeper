package app

import (
	"fmt"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/options"
	"github.com/spf13/cobra"
)

// newConfigCommand создаёт команду для настройки параметров клиента Gophkeeper.
// Позволяет сохранить токен аутентификации и/или URL сервера в переменных окружения или конфигурационных файлах.
// Для работы достаточно указать хотя бы один из параметров --token или --server-url.
func newConfigCommand() *cobra.Command {
	var (
		token     string
		serverURL string
	)

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
				if err := options.SetToken(token); err != nil {
					return fmt.Errorf("не удалось сохранить токен: %w", err)
				}
				fmt.Println("Токен сохранён в GOPHKEEPER_TOKEN")
			}

			if serverURL != "" {
				if err := options.SetServerURL(serverURL); err != nil {
					return fmt.Errorf("не удалось сохранить URL сервера: %w", err)
				}
				fmt.Println("URL сервера сохранён в GOPHKEEPER_SERVER_URL")
			}

			return nil
		},
	}

	cmd = options.RegisterTokenFlag(cmd, &token)
	cmd = options.RegisterServerURLFlag(cmd, &serverURL)

	return cmd
}
