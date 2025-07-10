package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	loginUsername    string // хранит имя пользователя для входа.
	loginPassword    string // хранит пароль пользователя для входа.
	loginServerURL   string // содержит URL сервера для входа.
	loginInteractive bool   // указывает, использовать ли интерактивный режим ввода при входе.
)

// newLoginCommand создаёт команду аутентификации пользователя.
func newLoginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Аутентификация пользователя",
		Long: `Команда для аутентификации пользователя в системе Gophkeeper.

Поддерживает передачу имени пользователя, пароля и URL сервера как через флаги,
так и через интерактивный ввод. Если URL сервера не передан через флаг, 
необходимо явно указать его в интерактивном режиме.

Использование:

  gophkeeper login --username alice --password secret123 --server-url https://example.com
  gophkeeper login --interactive
`,
		Example: `  gophkeeper login --username alice --password secret123 --server-url https://example.com
  gophkeeper login --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseLoginFlags(); err != nil {
				return err
			}

			// TODO: Реализовать логику аутентификации через клиент
			fmt.Printf("Authenticating user: %s with password: %s on server: %s\n",
				loginUsername, strings.Repeat("*", len(loginPassword)), loginServerURL)

			return nil
		},
	}

	cmd.Flags().StringVar(&loginUsername, "username", "", "Имя пользователя для аутентификации")
	cmd.Flags().StringVar(&loginPassword, "password", "", "Пароль пользователя")
	cmd.Flags().StringVar(&loginServerURL, "server-url", "", "URL сервера для аутентификации")
	cmd.Flags().BoolVar(&loginInteractive, "interactive", false, "Включить интерактивный режим ввода")

	return cmd
}

// parseLoginFlags обрабатывает флаги и интерактивный ввод для команды аутентификации.
//
// Проверяет, что обязательные поля username и password не пусты,
// а также что указан URL сервера.
func parseLoginFlags() error {
	if loginInteractive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Введите имя пользователя: ")
		userInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		loginUsername = strings.TrimSpace(userInput)

		fmt.Print("Введите пароль: ")
		passInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		loginPassword = strings.TrimSpace(passInput)

		fmt.Print("Введите URL сервера: ")
		urlInput, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		loginServerURL = strings.TrimSpace(urlInput)
	}

	if loginUsername == "" || loginPassword == "" {
		return fmt.Errorf("имя пользователя и пароль не могут быть пустыми")
	}

	if loginServerURL == "" {
		return fmt.Errorf("URL сервера должен быть указан")
	}

	return nil
}
