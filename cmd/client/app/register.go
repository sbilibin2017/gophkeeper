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
	registerUsername string // Имя пользователя для аутентификации
	registerPassword string // Пароль пользователя
)

// newRegisterCommand создаёт команду регистрации нового пользователя.
func newRegisterCommand() *cobra.Command {
	var (
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Регистрация нового пользователя",
		Long: `Команда регистрации нового пользователя в системе Gophkeeper.

Поддерживает передачу имени пользователя, пароля и URL сервера как через флаги,
так и через интерактивный ввод. Если URL сервера не передан, используется
значение из переменной окружения GOPHKEEPER_SERVER_URL.`,
		Example: `  gophkeeper register --username alice --password secret123 --server-url https://example.com
  gophkeeper register --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseRegisterFlags(&serverURL, &interactive); err != nil {
				return err
			}

			opts, err := options.NewOptions(
				options.WithServerURL(serverURL),
			)
			if err != nil {
				return fmt.Errorf("failed to create client config: %w", err)
			}
			if opts.ClientConfig.GRPCClient != nil {
				defer opts.ClientConfig.GRPCClient.Close()
			}

			fmt.Printf("Registering user: %s with password: %s at server: %s\n",
				registerUsername, strings.Repeat("*", len(registerPassword)), serverURL)

			return nil
		},
	}

	cmd.Flags().StringVar(&registerUsername, "username", "", "Имя пользователя для регистрации")
	cmd.Flags().StringVar(&registerPassword, "password", "", "Пароль пользователя")

	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	return cmd
}

// parseRegisterFlags обрабатывает флаги и интерактивный ввод для команды регистрации.
func parseRegisterFlags(serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseRegisterFlagsInteractive(reader, serverURL); err != nil {
			return err
		}
	}

	if registerUsername == "" || registerPassword == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	if *serverURL == "" {
		*serverURL = os.Getenv("GOPHKEEPER_SERVER_URL")
	}

	if *serverURL == "" {
		return fmt.Errorf("server-url must be provided")
	}

	return nil
}

// parseRegisterFlagsInteractive получает данные интерактивно.
func parseRegisterFlagsInteractive(r *bufio.Reader, serverURL *string) error {
	fmt.Print("Enter username: ")
	userInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	registerUsername = strings.TrimSpace(userInput)

	fmt.Print("Enter password: ")
	passInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	registerPassword = strings.TrimSpace(passInput)

	fmt.Print("Enter server URL (leave empty to use GOPHKEEPER_SERVER_URL env): ")
	urlInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*serverURL = strings.TrimSpace(urlInput)

	return nil
}
