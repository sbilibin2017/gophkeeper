package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/options"
	"github.com/spf13/cobra"
)

var (
	textData string         // текстовые данные (глобальная)
	textMeta flags.MetaFlag // метаданные (глобальная)
)

// newAddTextCommand создаёт команду для добавления произвольных текстовых данных с метаданными.
// Поддерживает передачу параметров через флаги и интерактивный ввод.
func newAddTextCommand() *cobra.Command {
	var (
		token       string
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "add-text",
		Short: "Добавить произвольные текстовые данные",
		Long: `Команда для добавления произвольных текстовых данных в систему Gophkeeper.

Позволяет передавать данные и метаданные как через флаги, так и через интерактивный ввод.
Обязательны параметры: текстовые данные, токен авторизации и URL сервера.

Пример использования:

  gophkeeper add-text --data "секретные заметки" --meta note=личное --token mytoken --server-url https://example.com
  gophkeeper add-text --interactive
  gophkeeper add-text --data "backup notes" --meta category=work --server-url https://example.com --token mytoken
`,
		Example: `  gophkeeper add-text --data "some secret text" --meta note=personal --token mytoken --server-url https://example.com
  gophkeeper add-text --interactive
  gophkeeper add-text --data "backup notes" --meta category=work --server-url https://example.com --token mytoken`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseTextFlags(&token, &serverURL, &interactive); err != nil {
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

			fmt.Printf("Добавлены данные:\nДанные: %s\nМетаданные: %+v\n", textData, textMeta)

			// TODO: реализовать сохранение текстовых данных с метаданными

			return nil
		},
	}

	cmd.Flags().StringVar(&textData, "data", "", "Текстовые данные для добавления")
	cmd.Flags().Var(&textMeta, "meta", "Метаданные в формате key=value (можно указывать несколько раз)")

	cmd = options.RegisterTokenFlag(cmd, &token)
	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	return cmd
}

// parseTextFlags обрабатывает флаги и интерактивный ввод для команды add-text.
// Проверяет обязательные параметры.
func parseTextFlags(token, serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseTextFlagsInteractive(reader, token, serverURL); err != nil {
			return err
		}
	}

	if textData == "" {
		return fmt.Errorf("параметр data обязателен")
	}
	if *token == "" || *serverURL == "" {
		return fmt.Errorf("токен и URL сервера должны быть заданы")
	}

	return nil
}

// parseTextFlagsInteractive запрашивает у пользователя необходимые параметры для добавления текстовых данных:
// текстовые данные, метаданные, токен и URL сервера.
func parseTextFlagsInteractive(r *bufio.Reader, token, serverURL *string) error {
	fmt.Print("Введите текстовые данные: ")
	inputData, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	textData = strings.TrimSpace(inputData)

	fmt.Println("Введите метаданные в формате key=value по одному, пустая строка — завершить:")
	for {
		fmt.Print("> ")
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if err := textMeta.Set(line); err != nil {
			return fmt.Errorf("некорректный ввод метаданных: %w", err)
		}
	}

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
