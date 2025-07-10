package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/config"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/spf13/cobra"
)

var (
	textData        string         // содержит текстовые данные для добавления.
	textToken       string         // хранит токен авторизации для запроса к серверу.
	textServerURL   string         // содержит URL сервера для отправки данных.
	textInteractive bool           // указывает, использовать ли интерактивный режим ввода.
	textMeta        flags.MetaFlag // содержит метаданные в формате ключ=значение.
)

// newAddTextCommand создаёт команду для добавления произвольного текстового данных с метаданными.
func newAddTextCommand() *cobra.Command {
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
			if err := parseTextFlags(); err != nil {
				return err
			}

			cfg, err := config.NewConfig(
				config.WithToken(textToken),
				config.WithServerURL(textServerURL),
			)
			if err != nil {
				return fmt.Errorf("не удалось создать конфигурацию клиента: %w", err)
			}
			if cfg.ClientConfig.GRPCClient != nil {
				defer cfg.ClientConfig.GRPCClient.Close()
			}

			fmt.Printf("Добавлены данные:\nДанные: %s\nТокен: %s\nURL сервера: %s\nМетаданные: %+v\n",
				textData,
				textToken,
				textServerURL,
				textMeta,
			)

			// TODO: реализовать сохранение текстовых данных с метаданными

			return nil
		},
	}

	cmd.Flags().StringVar(&textData, "data", "", "Текстовые данные для добавления")
	cmd.Flags().Var(&textMeta, "meta", "Метаданные в формате key=value (можно указывать несколько раз)")
	cmd.Flags().StringVar(&textToken, "token", "", "Токен авторизации")
	cmd.Flags().StringVar(&textServerURL, "server-url", "", "URL сервера")
	cmd.Flags().BoolVar(&textInteractive, "interactive", false, "Включить интерактивный режим ввода")

	return cmd
}

// parseTextFlags обрабатывает флаги и интерактивный ввод для команды add-text.
//
// Проверяет, что текстовые данные, токен и URL сервера заданы.
func parseTextFlags() error {
	if textInteractive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Введите текстовые данные: ")
		inputData, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		textData = strings.TrimSpace(inputData)

		fmt.Println("Введите метаданные в формате key=value по одному, пустая строка — завершить:")
		for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
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

		fmt.Print("Введите токен авторизации: ")
		inputToken, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		textToken = strings.TrimSpace(inputToken)

		fmt.Print("Введите URL сервера: ")
		inputServerURL, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		textServerURL = strings.TrimSpace(inputServerURL)
	}

	if textData == "" {
		return fmt.Errorf("параметр data обязателен")
	}
	if textToken == "" || textServerURL == "" {
		return fmt.Errorf("необходимо указать токен и URL сервера")
	}

	return nil
}
