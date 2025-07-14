package commands

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RegisterRegisterCommand добавляет в корневую команду подкоманду "register",
// которая выполняет регистрацию пользователя через HTTP или gRPC.
//
// Использование:
//
//	gophkeeper register --username=vasya --password=secret123 --auth-url=http://localhost:8080 --protocol=http
//
// Флаги:
//
//	--username  (обязательный) имя пользователя (только буквы и цифры)
//	--password  (обязательный) пароль (не менее 6 символов)
//	--auth-url  (обязательный) адрес сервера аутентификации
//	--protocol  протокол для связи с сервером: "http" (по умолчанию) или "grpc"
func RegisterRegisterCommand(root *cobra.Command) {
	var username string
	var password string
	var authURL string
	var protocol string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Регистрация пользователя",
		Long: `Производит регистрацию пользователя по имени и паролю.
После успешной регистрации клиент получает токен авторизации,
который можно использовать для дальнейших запросов.`,
		Example: `  gophkeeper register --username=vasya --password=secret123 --auth-url=http://localhost:8080 --protocol=http`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Проверка имени пользователя
			if username == "" {
				return fmt.Errorf("имя пользователя не может быть пустым")
			}

			matched, err := regexp.MatchString(`^[a-zA-Z0-9]+$`, username)
			if err != nil {
				return fmt.Errorf("ошибка при проверке имени пользователя")
			}
			if !matched {
				return fmt.Errorf("имя пользователя должно содержать только буквы и цифры (a-z, A-Z, 0-9)")
			}

			// Проверка пароля
			if len(password) < 6 {
				return fmt.Errorf("пароль должен содержать не менее 6 символов")
			}

			// Проверка адреса сервера
			if authURL == "" {
				return fmt.Errorf("адрес сервера не может быть пустым")
			}

			// Проверка протокола
			if protocol != "http" && protocol != "grpc" {
				return fmt.Errorf("протокол должен быть 'http' или 'grpc'")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var registerFacade interface {
				Register(context.Context, *models.AuthRequest) (string, error)
			}

			// Создание фасада в зависимости от протокола
			switch protocol {
			case "http":
				config, err := configs.NewClientConfig(
					configs.WithHTTPClient(authURL),
				)
				if err != nil {
					return fmt.Errorf("не удалось подключиться к HTTP серверу")
				}
				registerFacade = facades.NewRegisterHTTPFacade(config.HTTPClient)

			case "grpc":
				config, err := configs.NewClientConfig(
					configs.WithGRPCClient(authURL),
				)
				if err != nil {
					return fmt.Errorf("не удалось подключиться к gRPC серверу")
				}
				if config.GRPCClient == nil {
					return fmt.Errorf("подключение к gRPC серверу отсутствует")
				}
				grpcClient := pb.NewAuthServiceClient(config.GRPCClient)
				registerFacade = facades.NewRegisterGRPCFacade(grpcClient)
			}

			req := &models.AuthRequest{
				Username: username,
				Password: password,
			}

			// Вызов регистрации и вывод токена
			token, err := registerFacade.Register(ctx, req)
			if err != nil {
				// Возвращаем клиенту общее сообщение об ошибке без технических деталей
				return fmt.Errorf("не удалось зарегистрировать пользователя")
			}

			fmt.Println(token)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Имя пользователя (обязательный параметр)")
	cmd.Flags().StringVar(&password, "password", "", "Пароль (обязательный параметр)")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "URI сервера аутентификации и авторизации (обязательный параметр)")
	cmd.Flags().StringVar(&protocol, "protocol", "http", "Протокол для связи с сервером: 'http' или 'grpc'")

	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("auth-url")

	root.AddCommand(cmd)
}
