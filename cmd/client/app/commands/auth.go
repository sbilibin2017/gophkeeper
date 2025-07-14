package commands

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RegisterRegisterCommand добавляет в корневую команду подкоманду "register",
// которая производит регистрацию пользователя через HTTP или gRPC.
// Флаги:
//
//	--username  (обязательный) имя пользователя (только буквы и цифры)
//	--password  (обязательный) пароль (не менее 6 символов)
//	--auth-url  (обязательный) адрес сервера аутентификации
//	--protocol  протокол связи с сервером: "http" (по умолчанию) или "grpc"
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

			var config *configs.ClientConfig

			switch protocol {
			case models.ProtocolTypeHTTP, models.ProtocolTypeHTTPS:
				config, err = configs.NewClientConfig(
					configs.WithHTTPClient(authURL),
					configs.WithDB("gophkeeper_client.db"),
				)
				if err != nil {
					return fmt.Errorf("не удалось подключиться к HTTP серверу")
				}

			case models.ProtocolTypeGRPC:
				config, err = configs.NewClientConfig(
					configs.WithGRPCClient(authURL),
					configs.WithDB("gophkeeper_client.db"),
				)
				if err != nil {
					return fmt.Errorf("не удалось подключиться к gRPC серверу")
				}
				if config.GRPCClient == nil {
					return fmt.Errorf("подключение к gRPC серверу отсутствует")
				}
			}

			// Создаем все необходимые таблицы
			if err := createAllTables(config.DB); err != nil {
				return fmt.Errorf("внутренняя ошибка")
			}

			// Создаем фасад для регистрации в зависимости от протокола
			switch protocol {
			case models.ProtocolTypeHTTP, models.ProtocolTypeHTTPS:
				registerFacade = facades.NewRegisterHTTPFacade(config.HTTPClient)
			case models.ProtocolTypeGRPC:
				grpcClient := pb.NewAuthServiceClient(config.GRPCClient)
				registerFacade = facades.NewRegisterGRPCFacade(grpcClient)
			}

			req := &models.AuthRequest{
				Username: username,
				Password: password,
			}

			token, err := registerFacade.Register(ctx, req)
			if err != nil {
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

// RegisterLoginCommand добавляет в корневую команду подкоманду "login",
// которая выполняет аутентификацию пользователя через HTTP или gRPC.
//
// Использование:
//
//	gophkeeper login --username=vasya --password=secret123 --auth-url=http://localhost:8080 --protocol=http
//
// Флаги:
//
//	--username  (обязательный) имя пользователя (только буквы и цифры)
//	--password  (обязательный) пароль (не менее 6 символов)
//	--auth-url  (обязательный) адрес сервера аутентификации
//	--protocol  протокол для связи с сервером: "http" (по умолчанию) или "grpc"
func RegisterLoginCommand(root *cobra.Command) {
	var username string
	var password string
	var authURL string
	var protocol string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Аутентификация пользователя",
		Long: `Производит аутентификацию пользователя по имени и паролю.
После успешной аутентификации клиент получает токен авторизации,
который можно использовать для дальнейших запросов.`,
		Example: `  gophkeeper login --username=vasya --password=secret123 --auth-url=http://localhost:8080 --protocol=http`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Валидация username
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

			// Валидация пароля
			if len(password) < 6 {
				return fmt.Errorf("пароль должен содержать не менее 6 символов")
			}

			// Валидация адреса сервера
			if authURL == "" {
				return fmt.Errorf("адрес сервера не может быть пустым")
			}

			// Валидация протокола
			if protocol != "http" && protocol != "grpc" {
				return fmt.Errorf("протокол должен быть 'http' или 'grpc'")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var loginFacade interface {
				Login(context.Context, *models.AuthRequest) (string, error)
			}

			var config *configs.ClientConfig

			// Создание фасада в зависимости от протокола
			switch protocol {
			case models.ProtocolTypeHTTP, models.ProtocolTypeHTTPS:
				config, err = configs.NewClientConfig(
					configs.WithHTTPClient(authURL),
					configs.WithDB("gophkeeper_client.db"),
				)
				if err != nil {
					return fmt.Errorf("не удалось подключиться к HTTP серверу")
				}

			case models.ProtocolTypeGRPC:
				config, err = configs.NewClientConfig(
					configs.WithGRPCClient(authURL),
					configs.WithDB("gophkeeper_client.db"),
				)
				if err != nil {
					return fmt.Errorf("не удалось подключиться к gRPC серверу")
				}
				if config.GRPCClient == nil {
					return fmt.Errorf("подключение к gRPC серверу отсутствует")
				}
			}

			// Создаем все необходимые таблицы
			if err := createAllTables(config.DB); err != nil {
				return fmt.Errorf("внутренняя ошибка при создании таблиц: %w", err)
			}

			// Создаем фасад
			switch protocol {
			case models.ProtocolTypeHTTP, models.ProtocolTypeHTTPS:
				loginFacade = facades.NewLoginHTTPFacade(config.HTTPClient)
			case models.ProtocolTypeGRPC:
				grpcClient := pb.NewAuthServiceClient(config.GRPCClient)
				loginFacade = facades.NewLoginGRPCFacade(grpcClient)
			}

			req := &models.AuthRequest{
				Username: username,
				Password: password,
			}

			token, err := loginFacade.Login(ctx, req)
			if err != nil {
				return fmt.Errorf("не удалось выполнить аутентификацию пользователя")
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

// createAllTables создаёт все необходимые таблицы в базе данных клиента.
// Возвращает ошибку, если создание какой-либо таблицы не удалось.
func createAllTables(db *sqlx.DB) error {
	if err := createSecretBinaryClientTable(db); err != nil {
		return fmt.Errorf("ошибка создания таблицы secret_binary_client: %w", err)
	}
	if err := createSecretTextClientTable(db); err != nil {
		return fmt.Errorf("ошибка создания таблицы secret_text_client: %w", err)
	}
	if err := createSecretUsernamePasswordClientTable(db); err != nil {
		return fmt.Errorf("ошибка создания таблицы secret_username_password_client: %w", err)
	}
	if err := createSecretBankCardClientTable(db); err != nil {
		return fmt.Errorf("ошибка создания таблицы secret_bank_card_client: %w", err)
	}
	return nil
}

// createSecretBinaryClientTable удаляет, если существует, и создает таблицу для SecretBinaryClient
func createSecretBinaryClientTable(db *sqlx.DB) error {
	dropQuery := `DROP TABLE IF EXISTS secret_binary_client;`
	createQuery := `
	CREATE TABLE secret_binary_client (
		secret_name TEXT PRIMARY KEY,
		data BYTEA NOT NULL,
		meta TEXT NULL,
		updated_at TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(dropQuery); err != nil {
		return err
	}
	_, err := db.Exec(createQuery)
	return err
}

// createSecretTextClientTable удаляет и создает таблицу для SecretTextClient
func createSecretTextClientTable(db *sqlx.DB) error {
	dropQuery := `DROP TABLE IF EXISTS secret_text_client;`
	createQuery := `
	CREATE TABLE secret_text_client (
		secret_name TEXT PRIMARY KEY,
		content TEXT NOT NULL,
		meta TEXT NULL,
		updated_at TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(dropQuery); err != nil {
		return err
	}
	_, err := db.Exec(createQuery)
	return err
}

// createSecretUsernamePasswordClientTable удаляет и создает таблицу для SecretUsernamePasswordClient
func createSecretUsernamePasswordClientTable(db *sqlx.DB) error {
	dropQuery := `DROP TABLE IF EXISTS secret_username_password_client;`
	createQuery := `
	CREATE TABLE secret_username_password_client (
		secret_name TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		meta TEXT NULL,
		updated_at TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(dropQuery); err != nil {
		return err
	}
	_, err := db.Exec(createQuery)
	return err
}

// createSecretBankCardClientTable удаляет и создает таблицу для SecretBankCardClient
func createSecretBankCardClientTable(db *sqlx.DB) error {
	dropQuery := `DROP TABLE IF EXISTS secret_bank_card_client;`
	createQuery := `
	CREATE TABLE secret_bank_card_client (
		secret_name TEXT PRIMARY KEY,
		number TEXT NOT NULL,
		owner TEXT NULL,
		exp TEXT NOT NULL,
		cvv TEXT NOT NULL,
		meta TEXT NULL,
		updated_at TIMESTAMP NOT NULL
	);`

	if _, err := db.Exec(dropQuery); err != nil {
		return err
	}
	_, err := db.Exec(createQuery)
	return err
}
