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
	content string         // путь к бинарному файлу (глобальная переменная)
	meta    flags.MetaFlag // метаданные key=value (глобальная перемаенная)
)

// newAddBinaryCommand создаёт и возвращает команду "add-binary" для CLI.
// Команда позволяет добавить бинарные данные из файла с возможностью указать дополнительные метаданные в формате key=value.
// Поддерживается интерактивный режим, при котором параметры вводятся пошагово.
func newAddBinaryCommand() *cobra.Command {
	var (
		token       string
		serverURL   string
		interactive bool
	)

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
			if err := parseBinaryFlags(&token, &serverURL, &interactive); err != nil {
				return err
			}

			opts, err := options.NewOptions(
				options.WithToken(token),
				options.WithServerURL(serverURL),
			)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}

			if opts.ClientConfig.GRPCClient != nil {
				defer opts.ClientConfig.GRPCClient.Close()
			}

			fmt.Printf("Binary file added:\nFile: %s\nToken: %s\nServer URL: %s\nMetadata: %+v\n",
				content,
				token,
				serverURL,
				meta,
			)

			return nil
		},
	}

	cmd = options.RegisterTokenFlag(cmd, &token)
	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	cmd.Flags().StringVar(&content, "content", "", "Path to binary file")
	cmd.Flags().Var(&meta, "meta", "Metadata key=value pairs (can be specified multiple times)")

	return cmd
}

// parseBinaryFlags обрабатывает флаги команды add-binary.
// Если включён интерактивный режим, запрашивает необходимые параметры у пользователя.
// Проверяет обязательные параметры content, token и serverURL.
func parseBinaryFlags(token, serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseBinaryFlagsInteractive(reader, token, serverURL); err != nil {
			return err
		}
	}

	if content == "" {
		return fmt.Errorf("parameter content is required")
	}

	if *token == "" || *serverURL == "" {
		return fmt.Errorf("token and server URL must be provided")
	}

	return nil
}

// parseBinaryFlagsInteractive запрашивает у пользователя необходимые параметры для загрузки бинарного файла:
// путь к файлу, метаданные, токен авторизации и URL сервера.
func parseBinaryFlagsInteractive(r *bufio.Reader, token, serverURL *string) error {
	fmt.Print("Enter path to binary file: ")
	inputFile, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	content = strings.TrimSpace(inputFile)

	fmt.Println("Enter metadata key=value pairs one by one. Leave empty to finish:")
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
		if err := meta.Set(line); err != nil {
			return fmt.Errorf("invalid meta input: %w", err)
		}
	}

	fmt.Print("Enter authorization token (leave empty to use GOPHKEEPER_TOKEN environment variable): ")
	inputToken, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*token = strings.TrimSpace(inputToken)

	fmt.Print("Enter server URL (leave empty to use GOPHKEEPER_SERVER_URL environment variable): ")
	inputServerURL, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*serverURL = strings.TrimSpace(inputServerURL)

	return nil
}
