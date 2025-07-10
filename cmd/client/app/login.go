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
	loginUsername string // Имя пользователя для аутентификации
	loginPassword string // Пароль пользователя
)

// newLoginCommand создаёт команду для аутентификации пользователя в системе Gophkeeper.
// Поддерживает передачу имени пользователя, пароля и URL сервера через флаги или интерактивный ввод.
func newLoginCommand() *cobra.Command {
	var (
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Аутентификация пользователя",
		Long: `Команда для аутентификации пользователя в системе Gophkeeper.

Поддерживает передачу имени пользователя, пароля и URL сервера как через флаги,
так и через интерактивный ввод.`,
		Example: `  gophkeeper login --username alice --password secret123 --server-url https://example.com
  gophkeeper login --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseLoginFlags(&serverURL, &interactive); err != nil {
				return err
			}

			opts, err := options.NewOptions(
				options.WithServerURL(serverURL),
			)
			if err != nil {
				return fmt.Errorf("не удалось создать конфигурацию клиента: %w", err)
			}
			if opts.ClientConfig.GRPCClient != nil {
				defer opts.ClientConfig.GRPCClient.Close()
			}

			// TODO: реализовать вызов аутентификации пользователя через клиент
			fmt.Printf("Аутентификация пользователя: %s с паролем: %s на сервере: %s\n",
				loginUsername, strings.Repeat("*", len(loginPassword)), serverURL)

			return nil
		},
	}

	cmd.Flags().StringVar(&loginUsername, "username", "", "Имя пользователя для аутентификации")
	cmd.Flags().StringVar(&loginPassword, "password", "", "Пароль пользователя")

	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	return cmd
}

// parseLoginFlags обрабатывает флаги и интерактивный ввод для команды login.
func parseLoginFlags(serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseLoginFlagsInteractive(reader, serverURL); err != nil {
			return err
		}
	}

	if loginUsername == "" || loginPassword == "" {
		return fmt.Errorf("имя пользователя и пароль не могут быть пустыми")
	}

	if *serverURL == "" {
		return fmt.Errorf("URL сервера не может быть пустым")
	}

	return nil
}

// parseLoginFlagsInteractive запрашивает у пользователя имя, пароль и URL сервера.
func parseLoginFlagsInteractive(r *bufio.Reader, serverURL *string) error {
	fmt.Print("Введите имя пользователя: ")
	userInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	loginUsername = strings.TrimSpace(userInput)

	fmt.Print("Введите пароль: ")
	passInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	loginPassword = strings.TrimSpace(passInput)

	fmt.Print("Введите URL сервера: ")
	urlInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*serverURL = strings.TrimSpace(urlInput)

	return nil
}
