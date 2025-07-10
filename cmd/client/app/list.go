package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/options"
	"github.com/spf13/cobra"
)

var (
	secretType string
	outputType string
)

// newListCommand создаёт команду для отображения сохранённых данных.
// Поддерживает фильтрацию по типу секрета и выбор типа вывода (консоль, интерактивный, файл).
func newListCommand() *cobra.Command {
	var (
		token       string
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Отобразить сохранённые данные",
		Long: `Команда для отображения всех или отфильтрованных данных, сохранённых в системе Gophkeeper.

Поддерживает флаги и интерактивный режим для ввода токена, URL сервера и типа секрета.`,
		Example: `  gophkeeper list
  gophkeeper list --type login
  gophkeeper list --token mytoken --server-url https://example.com
  gophkeeper list --interactive
  gophkeeper list --output-type file
  gophkeeper list --output-type interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseListFlags(&token, &serverURL, &interactive); err != nil {
				return err
			}

			opts, err := options.NewOptions(
				options.WithToken(token),
				options.WithServerURL(serverURL),
			)
			if err != nil {
				return fmt.Errorf("не удалось создать конфигурацию клиента: %w", err)
			}
			if opts.ClientConfig.GRPCClient != nil {
				defer opts.ClientConfig.GRPCClient.Close()
			}

			// TODO: реализовать логику вывода списка секретов с учётом secretType и outputType

			return nil
		},
	}

	cmd.Flags().StringVar(&secretType, "type", "", "Фильтр по типу секрета (login, text, binary, card)")
	cmd.Flags().StringVar(&outputType, "output-type", "console", "Тип вывода результата (interactive, file, console)")

	cmd = options.RegisterTokenFlag(cmd, &token)
	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	return cmd
}

// parseListFlags обрабатывает флаги и интерактивный ввод для команды list.
func parseListFlags(token, serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseListFlagsInteractive(reader, token, serverURL); err != nil {
			return err
		}
	}

	if *token == "" || *serverURL == "" {
		return fmt.Errorf("необходимо указать токен и URL сервера (через флаги, интерактивно или переменные окружения)")
	}

	return nil
}

// parseListFlagsInteractive запрашивает у пользователя токен и URL сервера для команды list.
func parseListFlagsInteractive(r *bufio.Reader, token, serverURL *string) error {
	fmt.Print("Введите токен авторизации (оставьте пустым для использования GOPHKEEPER_TOKEN): ")
	inputToken, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*token = strings.TrimSpace(inputToken)

	fmt.Print("Введите URL сервера (оставьте пустым для использования GOPHKEEPER_SERVER_URL): ")
	inputServerURL, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*serverURL = strings.TrimSpace(inputServerURL)

	return nil
}
