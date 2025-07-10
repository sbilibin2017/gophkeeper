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
	resolver string // Стратегия разрешения конфликтов при синхронизации (server, client, interactive)
)

// newSyncCommand создаёт и возвращает команду "sync" для CLI.
func newSyncCommand() *cobra.Command {
	var (
		token       string
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Синхронизировать данные с сервером",
		Long: `Команда позволяет синхронизировать данные с сервером.

Обязательными параметрами являются токен авторизации и URL сервера.
Поддерживается интерактивный режим, при котором параметры вводятся пошагово.

Параметр resolver отвечает за стратегию разрешения конфликтов (server, client, interactive).

Примеры использования:

  gophkeeper sync --token mytoken --server-url https://example.com --resolver server
  gophkeeper sync --interactive
  gophkeeper sync --resolver interactive --token mytoken --server-url https://example.com
`,
		Example: `  gophkeeper sync --token mytoken --server-url https://example.com --resolver server
  gophkeeper sync --interactive
  gophkeeper sync --resolver interactive --token mytoken --server-url https://example.com`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseSyncFlags(&token, &serverURL, &resolver, &interactive); err != nil {
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

			fmt.Printf("Синхронизация с сервером:\nТокен: %s\nURL сервера: %s\nСтратегия разрешения конфликтов: %s\n",
				token, serverURL, resolver,
			)

			// TODO: Реализовать синхронизацию и разрешение конфликтов

			if resolver == "interactive" {
				fmt.Println("Активирован интерактивный режим разрешения конфликтов")
				// TODO: Реализовать интерактивное разрешение конфликтов
			}

			return nil
		},
	}

	cmd = options.RegisterTokenFlag(cmd, &token)
	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	cmd.Flags().StringVar(&resolver, "resolver", "", "Стратегия разрешения конфликтов (server, client, interactive)")

	return cmd
}

// parseSyncFlags обрабатывает флаги команды sync.
// В интерактивном режиме запрашивает ввод значений у пользователя.
// Проверяет обязательные параметры и возвращает ошибку при их отсутствии.
func parseSyncFlags(token, serverURL, resolver *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseSyncFlagsInteractive(reader, token, serverURL, resolver); err != nil {
			return err
		}
	}

	if *token == "" {
		*token = os.Getenv("GOPHKEEPER_TOKEN")
	}
	if *serverURL == "" {
		*serverURL = os.Getenv("GOPHKEEPER_SERVER_URL")
	}

	if *token == "" || *serverURL == "" {
		return fmt.Errorf("токен и URL сервера должны быть указаны")
	}

	validResolvers := map[string]bool{
		"server":      true,
		"client":      true,
		"interactive": true,
		"":            true, // разрешаем пустое значение
	}
	if !validResolvers[*resolver] {
		return fmt.Errorf("недопустимое значение resolver: %s. Допустимые значения: server, client, interactive", *resolver)
	}

	return nil
}

// parseSyncFlagsInteractive выполняет интерактивный ввод параметров команды sync.
func parseSyncFlagsInteractive(r *bufio.Reader, token, serverURL, resolver *string) error {
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

	fmt.Print("Введите стратегию разрешения конфликтов (server/client/interactive): ")
	inputResolver, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*resolver = strings.TrimSpace(inputResolver)

	return nil
}
