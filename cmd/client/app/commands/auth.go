package commands

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/protocol"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

func RegisterRegisterCommand(root *cobra.Command) {
	var username, password, authURL string

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Регистрация пользователя",
		Long:  `Регистрация пользователя и получение токена авторизации.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := "gophkeeper.db"
			if _, err := os.Stat(dbPath); err == nil {
				return fmt.Errorf("найден существующий клиент (%s); выполните синхронизацию или вход", dbPath)
			}

			if err := validateInput(username, password, authURL); err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			p := protocol.GetProtocolFromURL(authURL)

			cfg, err := prepareConfig(p, authURL, dbPath)
			if err != nil {
				return fmt.Errorf("не удалось подготовить конфигурацию клиента: %w", err)
			}

			if err := createAllTables(cfg.DB); err != nil {
				return fmt.Errorf("не удалось создать таблицы базы данных: %w", err)
			}

			req := models.AuthRequest{Username: username, Password: password}

			var token string
			switch p {
			case protocol.HTTP, protocol.HTTPS:
				token, err = client.RegisterUserHTTP(ctx, cfg.HTTPClient, &req)
			case protocol.GRPC:
				grpcClient := pb.NewAuthServiceClient(cfg.GRPCClient)
				token, err = client.RegisterUserGRPC(ctx, grpcClient, &req)
			default:
				return fmt.Errorf("неизвестный протокол: %s", p)
			}

			if err != nil {
				return fmt.Errorf("не удалось зарегистрировать пользователя: %w", err)
			}

			fmt.Println(token)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Имя пользователя (обязательно)")
	cmd.Flags().StringVar(&password, "password", "", "Пароль (обязательно)")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "URI сервера аутентификации (обязательно)")

	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("auth-url")

	root.AddCommand(cmd)
}

func RegisterLoginCommand(root *cobra.Command) {
	var username, password, authURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Аутентификация пользователя",
		Long:  `Аутентификация пользователя и получение токена авторизации.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			dbPath := "gophkeeper.db"
			if _, err := os.Stat(dbPath); err == nil {
				return fmt.Errorf("найден существующий клиент (%s); выполните синхронизацию или вход", dbPath)
			}

			if err := validateInput(username, password, authURL); err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			p := protocol.GetProtocolFromURL(authURL)

			cfg, err := prepareConfig(p, authURL, dbPath)
			if err != nil {
				return fmt.Errorf("не удалось подготовить конфигурацию клиента: %w", err)
			}

			if err := createAllTables(cfg.DB); err != nil {
				return fmt.Errorf("не удалось создать таблицы базы данных: %w", err)
			}

			req := models.AuthRequest{Username: username, Password: password}

			var token string
			switch p {
			case protocol.HTTP, protocol.HTTPS:
				token, err = client.LoginUserHTTP(ctx, cfg.HTTPClient, &req)
			case protocol.GRPC:
				grpcClient := pb.NewAuthServiceClient(cfg.GRPCClient)
				token, err = client.LoginUserGRPC(ctx, grpcClient, &req)
			default:
				return fmt.Errorf("неизвестный протокол: %s", p)
			}

			if err != nil {
				return fmt.Errorf("не удалось выполнить аутентификацию пользователя: %w", err)
			}

			fmt.Println(token)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "Имя пользователя (обязательно)")
	cmd.Flags().StringVar(&password, "password", "", "Пароль (обязательно)")
	cmd.Flags().StringVar(&authURL, "auth-url", "", "URI сервера аутентификации (обязательно)")

	_ = cmd.MarkFlagRequired("username")
	_ = cmd.MarkFlagRequired("password")
	_ = cmd.MarkFlagRequired("auth-url")

	root.AddCommand(cmd)
}

func validateInput(username, password, authURL string) error {
	if username == "" {
		return fmt.Errorf("имя пользователя не может быть пустым")
	}
	matched, err := regexp.MatchString(`^[a-zA-Z0-9]+$`, username)
	if err != nil || !matched {
		return fmt.Errorf("имя пользователя должно содержать только буквы и цифры")
	}
	if len(password) < 6 {
		return fmt.Errorf("пароль должен содержать не менее 6 символов")
	}
	if authURL == "" {
		return fmt.Errorf("адрес сервера не может быть пустым")
	}
	return nil
}

func prepareConfig(p, authURL, dbPath string) (*configs.ClientConfig, error) {
	switch p {
	case protocol.HTTP, protocol.HTTPS:
		return configs.NewClientConfig(
			configs.WithClientConfigHTTPClient(authURL),
			configs.WithClientConfigDB(dbPath),
		)
	case protocol.GRPC:
		return configs.NewClientConfig(
			configs.WithClientConfigGRPCClient(authURL),
			configs.WithClientConfigDB(dbPath),
		)
	default:
		return nil, fmt.Errorf("неизвестный протокол: %s", p)
	}
}

func createAllTables(db *sqlx.DB) error {
	if err := createSecretBinaryRequestTable(db); err != nil {
		return err
	}
	if err := createSecretTextRequestTable(db); err != nil {
		return err
	}
	if err := createSecretUsernamePasswordRequestTable(db); err != nil {
		return err
	}
	if err := createSecretBankCardRequestTable(db); err != nil {
		return err
	}
	return nil
}

func createSecretBinaryRequestTable(db *sqlx.DB) error {
	_, _ = db.Exec(`DROP TABLE IF EXISTS secret_binary_request;`)
	_, err := db.Exec(`
		CREATE TABLE secret_binary_request (
			secret_name TEXT PRIMARY KEY,
			data BYTEA NOT NULL,
			meta TEXT
		);
	`)
	return err
}

func createSecretTextRequestTable(db *sqlx.DB) error {
	_, _ = db.Exec(`DROP TABLE IF EXISTS secret_text_request;`)
	_, err := db.Exec(`
		CREATE TABLE secret_text_request (
			secret_name TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			meta TEXT
		);
	`)
	return err
}

func createSecretUsernamePasswordRequestTable(db *sqlx.DB) error {
	_, _ = db.Exec(`DROP TABLE IF EXISTS secret_username_password_request;`)
	_, err := db.Exec(`
		CREATE TABLE secret_username_password_request (
			secret_name TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			password TEXT NOT NULL,
			meta TEXT
		);
	`)
	return err
}

func createSecretBankCardRequestTable(db *sqlx.DB) error {
	_, _ = db.Exec(`DROP TABLE IF EXISTS secret_bank_card_request;`)
	_, err := db.Exec(`
		CREATE TABLE secret_bank_card_request (
			secret_name TEXT PRIMARY KEY,
			number TEXT NOT NULL,
			owner TEXT,
			exp TEXT NOT NULL,
			cvv TEXT NOT NULL,
			meta TEXT
		);
	`)
	return err
}
