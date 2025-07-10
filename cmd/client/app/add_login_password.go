package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/options"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

var (
	loginPasswordUsername string         // имя пользователя (глобальная)
	loginPasswordPassword string         // пароль пользователя (глобальная)
	loginPasswordMeta     flags.MetaFlag // метаданные (глобальная)
)

// newAddLoginPasswordCommand создаёт команду для добавления пары логин/пароль с метаданными.
// Команда поддерживает передачу параметров через флаги и интерактивный ввод.
func newAddLoginPasswordCommand() *cobra.Command {
	var (
		token       string
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "add-login-password",
		Short: "Добавить логин и пароль с опциональными метаданными",
		Long: `Команда позволяет добавить в систему пару логин/пароль с возможностью
указать дополнительные метаданные в формате key=value.

Параметры username и password обязательны для заполнения. Также необходимы
токен авторизации и URL сервера.

Поддерживается интерактивный режим для удобного ввода данных.

Пример использования:

  gophkeeper add-login-password --username user123 --password secret --meta site=example.com --token mytoken --server-url https://example.com
  gophkeeper add-login-password --interactive`,
		Example: `  gophkeeper add-login-password --username user123 --password secret --meta site=example.com --token mytoken --server-url https://example.com
  gophkeeper add-login-password --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseLoginPasswordFlags(&token, &serverURL, &interactive); err != nil {
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

			model := models.LoginPassword{
				Username: loginPasswordUsername,
				Password: loginPasswordPassword,
				Meta:     loginPasswordMeta,
			}

			fmt.Printf("Добавлены данные логина и пароля:\nЛогин: %s\nПароль: %s\nМетаданные: %+v\n",
				model.Username,
				model.Password,
				model.Meta,
			)

			// TODO: Реализовать сохранение логина/пароля на сервер

			return nil
		},
	}

	cmd.Flags().StringVar(&loginPasswordUsername, "username", "", "Логин пользователя")
	cmd.Flags().StringVar(&loginPasswordPassword, "password", "", "Пароль пользователя")
	cmd.Flags().Var(&loginPasswordMeta, "meta", "Метаданные в формате key=value (можно указывать несколько раз)")

	cmd = options.RegisterTokenFlag(cmd, &token)
	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	return cmd
}

// parseLoginPasswordFlags обрабатывает флаги и интерактивный ввод для команды add-login-password.
// Проверяет обязательные параметры и возвращает ошибку при их отсутствии.
func parseLoginPasswordFlags(token, serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseLoginPasswordFlagsInteractive(reader, token, serverURL); err != nil {
			return err
		}
	}

	if loginPasswordUsername == "" || loginPasswordPassword == "" {
		return fmt.Errorf("параметры username и password обязательны для заполнения")
	}
	if *token == "" || *serverURL == "" {
		return fmt.Errorf("токен и URL сервера должны быть заданы")
	}

	return nil
}

// parseLoginPasswordFlagsInteractive запрашивает у пользователя необходимые параметры для добавления логина и пароля:
// логин, пароль, метаданные, токен и URL сервера.
func parseLoginPasswordFlagsInteractive(r *bufio.Reader, token, serverURL *string) error {
	fmt.Print("Введите логин: ")
	inputLogin, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	loginPasswordUsername = strings.TrimSpace(inputLogin)

	fmt.Print("Введите пароль: ")
	inputPassword, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	loginPasswordPassword = strings.TrimSpace(inputPassword)

	fmt.Println("Введите метаданные в формате key=value по одному. Пустая строка — завершить:")
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
		if err := loginPasswordMeta.Set(line); err != nil {
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
