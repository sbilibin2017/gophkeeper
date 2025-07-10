package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/config"
	"github.com/spf13/cobra"
)

var (
	registerUsername    string // хранит имя пользователя для регистрации.
	registerPassword    string // хранит пароль пользователя для регистрации.
	registerServerURL   string // содержит URL сервера для регистрации.
	registerInteractive bool   // указывает, использовать ли интерактивный режим ввода при регистрации.
)

// newRegisterCommand создаёт команду регистрации нового пользователя.
func newRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Регистрация нового пользователя",
		Long: `Команда регистрации нового пользователя в системе Gophkeeper.

Поддерживает передачу имени пользователя, пароля и URL сервера как через флаги,
так и через интерактивный ввод. Если URL сервера не передан, используется
значение из переменной окружения GOPHKEEPER_SERVER_URL.

После сбора данных формируется конфигурация клиента, которая будет использоваться
для взаимодействия с сервером при регистрации.

Использование:

  gophkeeper register --username alice --password secret123 --server-url https://example.com
  gophkeeper register --interactive
`,
		Example: `  gophkeeper register --username alice --password secret123 --server-url https://example.com
  gophkeeper register --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseRegisterFlags(); err != nil {
				return err
			}

			cfg, err := config.NewConfig(
				config.WithServerURL(registerServerURL),
			)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}
			if cfg.ClientConfig.GRPCClient != nil {
				defer cfg.ClientConfig.GRPCClient.Close()
			}

			// TODO: Реализовать вызов регистрации пользователя через клиент
			fmt.Printf("Registering user: %s with password: %s at server: %s\n",
				registerUsername, strings.Repeat("*", len(registerPassword)), registerServerURL)

			return nil
		},
	}

	cmd.Flags().StringVar(&registerUsername, "username", "", "Имя пользователя для регистрации")
	cmd.Flags().StringVar(&registerPassword, "password", "", "Пароль пользователя")
	cmd.Flags().StringVar(&registerServerURL, "server-url", "", "URL сервера для регистрации")
	cmd.Flags().BoolVar(&registerInteractive, "interactive", false, "Включить интерактивный режим ввода")

	return cmd
}

// parseRegisterFlags обрабатывает флаги и интерактивный ввод для команды регистрации.
//
// Проверяет, что обязательные поля username и password не пусты,
// а также что указан URL сервера или он есть в окружении GOPHKEEPER_SERVER_URL.
func parseRegisterFlags() error {
	if registerInteractive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter username: ")
		userInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		registerUsername = strings.TrimSpace(userInput)

		fmt.Print("Enter password: ")
		passInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		registerPassword = strings.TrimSpace(passInput)

		fmt.Print("Enter server URL (leave empty to use GOPHKEEPER_SERVER_URL env): ")
		urlInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		registerServerURL = strings.TrimSpace(urlInput)
	}

	if registerUsername == "" || registerPassword == "" {
		return fmt.Errorf("username and password cannot be empty")
	}
	if registerServerURL == "" && os.Getenv("GOPHKEEPER_SERVER_URL") == "" {
		return fmt.Errorf("server-url must be provided")
	}

	return nil
}
