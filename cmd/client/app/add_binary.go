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
	binaryToken       string         // хранит токен авторизации для запроса к серверу.
	binaryServerURL   string         // хранит URL сервера, на который отправляются данные.
	binaryInteractive bool           // указывает, использовать ли интерактивный режим ввода.
	binaryContent     string         // содержит путь к бинарному файлу.
	binaryMeta        flags.MetaFlag // одержит метаданные в формате ключ=значение.
)

// newAddBinaryCommand создаёт и возвращает команду "add-binary" для CLI.
// Команда позволяет добавить бинарные данные из файла с опциональными метаданными.
func newAddBinaryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-binary",
		Short: "Add binary data from file with optional text metadata",
		Long: `Команда позволяет добавить бинарные данные из файла с возможностью указать дополнительные метаданные в формате key=value.

Обязательным параметром является путь к бинарному файлу.
Также требуется указать токен авторизации и URL сервера для отправки данных.

Поддерживается интерактивный режим, при котором все параметры вводятся пошагово.

Примеры использования:

  gophkeeper add-binary --content ./path/to/file.bin --meta site=example.com --meta user=john --token mytoken --server-url https://example.com
  gophkeeper add-binary --interactive
  gophkeeper add-binary --content backup.bin --meta codes="1234,5678,9012" --server-url https://example.com --token mytoken
`,
		Example: `  gophkeeper add-binary --content ./path/to/file.bin --meta site=example.com --meta user=john --token mytoken --server-url https://example.com
  gophkeeper add-binary --interactive
  gophkeeper add-binary --content backup.bin --meta codes="1234,5678,9012" --server-url https://example.com --token mytoken`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Парсим входные параметры команды (флаги).
			if err := parseBinaryFlags(); err != nil {
				return err
			}

			// Создаём конфигурацию клиента с токеном и URL сервера.
			cfg, err := config.NewConfig(
				config.WithToken(binaryToken),
				config.WithServerURL(binaryServerURL),
			)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			// Закрываем gRPC клиент после завершения функции, если он есть.
			if cfg.ClientConfig.GRPCClient != nil {
				defer cfg.ClientConfig.GRPCClient.Close()
			}

			// Для демонстрации выводим данные на экран.
			fmt.Printf("Binary file added:\nFile: %s\nToken: %s\nServer URL: %s\nMetadata: %+v\n",
				binaryContent,
				binaryToken,
				binaryServerURL,
				binaryMeta,
			)

			return nil
		},
	}

	// Определяем флаги команды.
	cmd.Flags().StringVar(&binaryToken, "token", "", "Authorization token (can be set via GOPHKEEPER_TOKEN env variable)")
	cmd.Flags().StringVar(&binaryServerURL, "server-url", "", "Server URL (can be set via GOPHKEEPER_SERVER_URL env variable)")
	cmd.Flags().BoolVar(&binaryInteractive, "interactive", false, "Interactive input mode")
	cmd.Flags().StringVar(&binaryContent, "content", "", "Path to binary file")
	cmd.Flags().Var(&binaryMeta, "meta", "Metadata key=value pairs (can be specified multiple times)")

	return cmd
}

// parseBinaryFlags обрабатывает флаги команды.
// В интерактивном режиме запрашивает ввод значений у пользователя.
// Проверяет обязательные параметры и возвращает ошибку при их отсутствии.
func parseBinaryFlags() error {
	if binaryInteractive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter path to binary file: ")
		inputFile, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		binaryContent = strings.TrimSpace(inputFile)

		fmt.Println("Enter metadata key=value pairs one by one. Leave empty to finish:")
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
			if err := binaryMeta.Set(line); err != nil {
				return fmt.Errorf("invalid meta input: %w", err)
			}
		}

		fmt.Print("Enter authorization token (leave empty to use GOPHKEEPER_TOKEN environment variable): ")
		inputToken, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		binaryToken = strings.TrimSpace(inputToken)

		fmt.Print("Enter server URL (leave empty to use GOPHKEEPER_SERVER_URL environment variable): ")
		inputServerURL, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		binaryServerURL = strings.TrimSpace(inputServerURL)
	}

	if binaryContent == "" {
		return fmt.Errorf("parameter content is required")
	}
	if binaryToken == "" || binaryServerURL == "" {
		return fmt.Errorf("token and server URL must be provided")
	}

	return nil
}
