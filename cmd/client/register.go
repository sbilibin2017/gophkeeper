package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/spf13/cobra"
)

var (
	registerServerURL   string // хранит URL сервера GophKeeper, к которому будет отправлен запрос регистрации.
	registerInteractive bool   // указывает, следует ли запрашивать логин и пароль интерактивно через консоль.
)

// init инициализирует флаги команды register.
func init() {
	registerCmd.Flags().StringVar(&registerServerURL, "server-url", "http://localhost:8080", "URL сервера GophKeeper")
	registerCmd.Flags().BoolVar(&registerInteractive, "interactive", false, "Запросить ввод логина и пароля в интерактивном режиме")
}

// registerCmd команда для регистрации нового пользователя.
var registerCmd = &cobra.Command{
	Use:   "register [login] [password]",
	Short: "Зарегистрировать нового пользователя",
	Example: `
		gophkeeper register user@example.com mypassword
		gophkeeper register user@example.com mypassword --server-url http://example.com
		gophkeeper register --interactive
`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var secret *models.UsernamePassword
		var err error

		// 1. Получаем логин и пароль:
		//    - из интерактивного ввода, если установлен флаг --interactive,
		//    - либо из аргументов командной строки.
		if registerInteractive {
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Введите логин: ")
			inputLogin, err := reader.ReadString('\n')
			if err != nil {
				return errors.New("ошибка при вводе логина")
			}
			login := strings.TrimSpace(inputLogin)

			fmt.Print("Введите пароль: ")
			inputPassword, err := reader.ReadString('\n')
			if err != nil {
				return errors.New("ошибка при вводе пароля")
			}
			password := strings.TrimSpace(inputPassword)

			secret = &models.UsernamePassword{
				Username: login,
				Password: password,
			}
		} else {
			if len(args) != 2 {
				return errors.New("неверный формат аргументов для логина и пароля")
			}
			secret = &models.UsernamePassword{
				Username: args[0],
				Password: args[1],
			}
		}

		// 2. Валидируем логин и пароль: они не должны быть пустыми.
		if secret.Username == "" || secret.Password == "" {
			return errors.New("логин или пароль не могут быть пустыми")
		}

		// 3. Определяем тип клиента (HTTP или gRPC) на основе схемы URL сервера.
		var opts []configs.ClientConfigOpt
		switch {
		case strings.HasPrefix(registerServerURL, "http://"), strings.HasPrefix(registerServerURL, "https://"):
			opts = append(opts, configs.WithHTTPClient(registerServerURL))
		case strings.HasPrefix(registerServerURL, "grpc://"):
			opts = append(opts, configs.WithGRPCClient(registerServerURL))
		}

		// 4. Создаём клиентскую конфигурацию.
		config, err := configs.NewClientConfig(opts...)
		if err != nil {
			return errors.New("не удалось создать конфигурацию клиента")
		}

		// 5. Создаём фасад для клиента
		var facade interface {
			Register(ctx context.Context, secret *models.UsernamePassword) (string, error)
		}

		switch {
		case config.HTTPClient != nil:
			facade, err = facades.NewRegisterHTTPFacade(config.HTTPClient)
			if err != nil {
				return errors.New("не удалось подключиться к серверу по HTTP")
			}
		case config.GRPCClient != nil:
			facade, err = facades.NewRegisterGRPCFacade(config.GRPCClient)
			if err != nil {
				return errors.New("не удалось подключиться к серверу по gRPC")
			}
		default:
			return errors.New("нет доступного клиента для подключения")
		}

		// 6. Создаем сервис для регистрации пользователя
		service := services.NewRegisterService(facade)

		// 7. Выполняем регистрацию и получаем JWT токен.
		token, err := service.Register(context.Background(), secret)
		if err != nil {
			return errors.New("ошибка регистрации пользователя")
		}

		// 8. Сохраняем URL сервера и токен в переменные окружения.
		err = os.Setenv(serverURLEnvKey, registerServerURL)
		if err != nil {
			return errors.New("не удалось установить переменную окружения для адреса сервера")
		}
		err = os.Setenv(tokenEnvKey, token)
		if err != nil {
			return errors.New("не удалось установить переменную окружения для токена")
		}

		return nil
	},
}
