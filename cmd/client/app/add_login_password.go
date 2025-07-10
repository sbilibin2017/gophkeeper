package app

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/config"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/flags"

	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
)

var (
	loginPasswordUsername    string         // содержит имя пользователя для входа с паролем.
	loginPasswordPassword    string         // содержит пароль пользователя.
	loginPasswordToken       string         // хранит токен авторизации для запроса к серверу.
	loginPasswordServerURL   string         // содержит URL сервера для отправки данных.
	loginPasswordInteractive bool           // указывает, использовать ли интерактивный режим ввода.
	loginPasswordMeta        flags.MetaFlag // содержит метаданные в формате ключ=значение.
)

// newAddLoginPasswordCommand создаёт команду для добавления пары логин/пароль с метаданными.
func newAddLoginPasswordCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-login-password",
		Short: "Добавить логин и пароль с опциональными метаданными",
		Long: `Команда позволяет добавить в систему пару логин/пароль с возможностью
указать дополнительные метаданные в формате key=value.

Параметры username и password обязательны для заполнения. Также необходимы
токен авторизации и URL сервера, которые можно передать через флаги или
задать через переменные окружения GOPHKEEPER_TOKEN и GOPHKEEPER_SERVER_URL.

Поддерживается интерактивный режим, позволяющий вводить данные пошагово.

Пример использования:

  gophkeeper add-login-password --username user123 --password secret --meta site=example.com --meta type=personal --token mytoken --server-url https://example.com
  gophkeeper add-login-password --interactive
`,
		Example: `  gophkeeper add-login-password --username user123 --password secret --meta site=example.com --meta type=personal --token mytoken --server-url https://example.com
  gophkeeper add-login-password --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseLoginPasswordFlags(); err != nil {
				return err
			}

			cfg, err := config.NewConfig(
				config.WithToken(loginPasswordToken),
				config.WithServerURL(loginPasswordServerURL),
			)
			if err != nil {
				return fmt.Errorf("не удалось создать конфигурацию клиента: %w", err)
			}
			if cfg.ClientConfig.GRPCClient != nil {
				defer cfg.ClientConfig.GRPCClient.Close()
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

	cmd.Flags().BoolVar(&loginPasswordInteractive, "interactive", false, "Включить интерактивный режим ввода")

	cmd.Flags().StringVar(&loginPasswordToken, "token", "", "Токен авторизации")
	cmd.Flags().StringVar(&loginPasswordServerURL, "server-url", "", "URL сервера")

	cmd.Flags().StringVar(&loginPasswordUsername, "username", "", "Логин пользователя")
	cmd.Flags().StringVar(&loginPasswordPassword, "password", "", "Пароль пользователя")
	cmd.Flags().Var(&loginPasswordMeta, "meta", "Метаданные в формате key=value (можно указывать несколько раз)")

	return cmd
}

// parseLoginPasswordFlags обрабатывает флаги и интерактивный ввод для команды add-login-password.
//
// Проверяет, что username и password указаны, а также token и server-url переданы
// либо через флаги, либо через переменные окружения.
func parseLoginPasswordFlags() error {
	if loginPasswordInteractive {
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Введите логин: ")
		inputLogin, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		loginPasswordUsername = strings.TrimSpace(inputLogin)

		fmt.Print("Введите пароль: ")
		inputPassword, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		loginPasswordPassword = strings.TrimSpace(inputPassword)

		fmt.Println("Введите метаданные в формате key=value по одному. Пустая строка — завершить:")
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
			if err := loginPasswordMeta.Set(line); err != nil {
				return fmt.Errorf("некорректный ввод метаданных: %w", err)
			}
		}

		fmt.Print("Введите токен авторизации (оставьте пустым для использования GOPHKEEPER_TOKEN): ")
		inputToken, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		loginPasswordToken = strings.TrimSpace(inputToken)
		if loginPasswordToken == "" {
			loginPasswordToken = os.Getenv("GOPHKEEPER_TOKEN")
		}

		fmt.Print("Введите URL сервера (оставьте пустым для использования GOPHKEEPER_SERVER_URL): ")
		inputServerURL, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		loginPasswordServerURL = strings.TrimSpace(inputServerURL)
		if loginPasswordServerURL == "" {
			loginPasswordServerURL = os.Getenv("GOPHKEEPER_SERVER_URL")
		}
	}

	if loginPasswordUsername == "" || loginPasswordPassword == "" {
		return fmt.Errorf("параметры username и password обязательны для заполнения")
	}
	if loginPasswordToken == "" || loginPasswordServerURL == "" {
		return fmt.Errorf("необходимо указать токен и URL сервера (через флаги или переменные окружения)")
	}

	return nil
}
