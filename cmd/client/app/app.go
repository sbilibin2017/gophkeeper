package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/sbilibin2017/gophkeeper/internal/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
	"github.com/spf13/cobra"
)

// NewRootCommand создает корневую CLI-команду для клиента GophKeeper.
func NewRootCommand() *cobra.Command {
	return &cobra.Command{Use: "gophkeeper-client"}
}

// NewRegisterCommand создает CLI-команду для регистрации пользователя и устройства.
func NewRegisterCommand() *cobra.Command {
	var (
		serverURL string
		username  string
		password  string
	)

	cmd := &cobra.Command{
		Use:     "register",
		Short:   "Регистрация нового пользователя и устройства GophKeeper",
		Example: "gophkeeper-client register --username user1 --password secret123 --server-url http://localhost:8080",
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID, err := getDeviceID()
			if err != nil {
				return err
			}
			privKey, token, err := runRegisterHTTP(
				cmd.Context(),
				serverURL,
				"migrations",
				username,
				password,
				deviceID,
			)
			if err != nil {
				return err
			}

			cmd.Println("Регистрация успешна")
			cmd.Printf("Приватный ключ: %s\n", string(privKey))
			cmd.Printf("Токен: %s\n", token)

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "URL сервера GophKeeper")
	cmd.Flags().StringVar(&username, "username", "", "имя пользователя для регистрации")
	cmd.Flags().StringVar(&password, "password", "", "пароль для регистрации")

	return cmd
}

// runRegisterHTTP подключается к базе данных пользователя, инициализирует HTTP-клиент
// и выполняет регистрацию нового пользователя и устройства через AuthHTTPFacade.
func runRegisterHTTP(
	ctx context.Context,
	serverURL string,
	databaseMigrationsDir string,
	username string,
	password string,
	deviceID string,
) ([]byte, string, error) {
	conn, err := db.New(
		"sqlite",
		fmt.Sprintf("%s.db", username),
	)
	if err != nil {
		return nil, "", err
	}
	defer conn.Close()

	goose.SetDialect("sqlite")
	if databaseMigrationsDir != "" {
		if err := goose.Up(conn.DB, databaseMigrationsDir); err != nil {
			return nil, "", err
		}
	}

	client := http.New(serverURL).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	authFacade := facades.NewAuthHTTPFacade(client, jwt.GetTokenFromRestyResponse)

	privKey, token, err := authFacade.Register(ctx, username, password, deviceID)
	if err != nil {
		return nil, "", err
	}

	return privKey, token, nil
}

// getDeviceID возвращает уникальный идентификатор устройства.
// Если идентификатор ранее был сохранён в файле, он читается.
// Если файла нет — создается новый UUID и сохраняется.
func getDeviceID() (string, error) {
	// файл для хранения ID (в текущей директории)
	filePath := filepath.Join(".", ".device_id")

	// Проверяем, есть ли сохранённый ID
	if data, err := os.ReadFile(filePath); err == nil {
		id := string(data)
		if id != "" {
			return id, nil
		}
	}

	// Генерируем новый UUID
	newID := uuid.New().String()

	// Сохраняем его в файл
	if err := os.WriteFile(filePath, []byte(newID), 0644); err != nil {
		return "", errors.New("не удалось сохранить deviceID: " + err.Error())
	}

	return newID, nil
}
