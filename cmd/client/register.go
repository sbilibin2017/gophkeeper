package main

import (
	"context"
	"errors"
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
	secretKey           string // хранит секретный ключ для генерации JWT (не используется в данном коде, но объявлен).
)

// init инициализирует флаги команды register.
func init() {
	registerCmd.Flags().StringVar(&registerServerURL, "server-url", "http://localhost:8080", "URL сервера GophKeeper")
	registerCmd.Flags().BoolVar(&registerInteractive, "interactive", false, "Запросить ввод логина и пароля в интерактивном режиме")
	registerCmd.Flags().StringVar(&secretKey, "secret-key", "", "Секретный ключ для генерации JWT")
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
		// 1. Создаём контекст для выполнения запроса регистрации.
		ctx := context.Background()

		var secret *models.UsernamePassword
		var err error

		// 2. Получаем логин и пароль:
		//    - из интерактивного ввода, если установлен флаг --interactive,
		//    - либо из аргументов командной строки.
		if registerInteractive {
			secret, err = models.NewUsernamePasswordFromInteractive(os.Stdin)
			if err != nil {
				return errors.New("ошибка при вводе логина и пароля")
			}
		} else {
			secret, err = models.NewUsernamePasswordFromArgs(args)
			if err != nil {
				return errors.New("неверный формат аргументов для логина и пароля")
			}
		}

		// 3. Валидируем полученный логин и пароль.
		if err = models.ValidateUsernamePassword(secret); err != nil {
			return errors.New("логин или пароль невалидны")
		}

		// 4. Определяем тип клиента (HTTP или gRPC) на основе схемы URL сервера.
		var opts []configs.ClientConfigOpt
		switch {
		case strings.HasPrefix(registerServerURL, "http://"), strings.HasPrefix(registerServerURL, "https://"):
			opts = append(opts, configs.WithHTTPClient(registerServerURL))
		case strings.HasPrefix(registerServerURL, "grpc://"):
			opts = append(opts, configs.WithGRPCClient(registerServerURL))
		default:
			return errors.New("неподдерживаемая схема URL сервера")
		}

		// 5. Создаём клиентскую конфигурацию с выбранными параметрами.
		config, err := configs.NewClientConfig(opts...)
		if err != nil {
			return errors.New("не удалось создать конфигурацию клиента")
		}

		var service *services.RegisterService

		// 6. Создаём фасад и сервис регистрации для HTTP клиента, если он есть.
		if config.HTTPClient != nil {
			facade, err := facades.NewRegisterHTTPFacade(config.HTTPClient)
			if err != nil {
				return errors.New("не удалось подключиться к серверу по HTTP")
			}
			service, err = services.NewRegisterService(facade)
			if err != nil {
				return errors.New("ошибка создания сервиса регистрации")
			}
		}

		// 7. Аналогично создаём фасад и сервис для gRPC клиента.
		if config.GRPCClient != nil {
			facade, err := facades.NewRegisterGRPCFacade(config.GRPCClient)
			if err != nil {
				return errors.New("не удалось подключиться к серверу по gRPC")
			}
			service, err = services.NewRegisterService(facade)
			if err != nil {
				return errors.New("ошибка создания сервиса регистрации")
			}
		}

		// 8. Выполняем регистрацию и получаем JWT токен.
		token, err := service.Register(ctx, secret)
		if err != nil {
			return errors.New("ошибка регистрации пользователя")
		}

		// 9. Сохраняем URL сервера и полученный токен в переменные окружения.
		configs.SetServerURLToEnv(registerServerURL)
		configs.SetTokenToEnv(token)

		// 10. Завершаем команду без ошибок.
		return nil
	},
}
