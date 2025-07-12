package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/config"
	"github.com/sbilibin2017/gophkeeper/cmd/client/app/commands/parsemeta"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Auth выполняет аутентификацию пользователя.
// Принимает контекст ctx, аргументы args, флаги flags, переменные окружения envs и reader для интерактивного ввода.
// Возвращает ошибку в случае неудачи.
func Auth(
	ctx context.Context,
	args []string,
	flags map[string]string,
	envs []string,
	reader *bufio.Reader,
) error {
	var (
		secret *models.UsernamePassword
		err    error
	)

	// Парсинг параметров
	serverURL, interactive, err := parseAuthFlags(flags)
	if err != nil {
		return err
	}

	if interactive {
		secret, err = parseAuthFlagsInteractive(reader)
	} else {
		secret, err = parseAuthArgs(args)
	}
	if err != nil {
		return err
	}

	if err = validateAuthRequest(secret); err != nil {
		return err
	}

	// Создание конфигурации клиента по URL сервера.
	cfg, err := config.NewConfig(serverURL)
	if err != nil {
		return err
	}

	// Запуск аутентификации через HTTP или gRPC.
	token, err := runAuth(ctx, cfg, secret)
	if err != nil {
		return err
	}

	// Установка переменных окружения с данными аутентификации.
	if err = setAuthEnv(serverURL, token); err != nil {
		return err
	}

	return nil
}

// parseAuthFlags разбирает флаги команды для получения serverURL и режима interactive.
// Возвращает URL сервера, флаг интерактивного режима и ошибку, если флаги некорректны.
func parseAuthFlags(flags map[string]string) (string, bool, error) {
	var err error

	serverURL := ""
	if v, ok := flags["server-url"]; ok && v != "" {
		serverURL = v
	}

	interactive := false
	if v, ok := flags["interactive"]; ok && v != "" {
		interactive, err = strconv.ParseBool(v)
		if err != nil {
			return "", false, errors.New("некорректное значение флага --interactive")
		}
	}

	return serverURL, interactive, nil
}

// parseAuthFlagsInteractive запрашивает у пользователя логин, пароль и метаданные через reader (например, stdin).
// Возвращает заполненную структуру UsernamePassword или ошибку при некорректном вводе.
func parseAuthFlagsInteractive(reader *bufio.Reader) (*models.UsernamePassword, error) {
	fmt.Print("Введите логин: ")
	inputLogin, err := reader.ReadString('\n')
	if err != nil {
		return nil, errors.New("ошибка при вводе логина")
	}
	login := strings.TrimSpace(inputLogin)
	if login == "" {
		return nil, errors.New("логин не может быть пустым")
	}

	fmt.Print("Введите пароль: ")
	inputPassword, err := reader.ReadString('\n')
	if err != nil {
		return nil, errors.New("ошибка при вводе пароля")
	}
	password := strings.TrimSpace(inputPassword)
	if password == "" {
		return nil, errors.New("пароль не может быть пустым")
	}

	meta, err := parsemeta.ParseMetaInteractive(reader)
	if err != nil {
		return nil, err
	}

	return &models.UsernamePassword{
		Username: login,
		Password: password,
		Meta:     meta,
	}, nil
}

// parseAuthArgs проверяет переданные аргументы командной строки и возвращает структуру UsernamePassword.
// Если аргументов недостаточно, возвращает ошибку.
func parseAuthArgs(args []string) (*models.UsernamePassword, error) {
	if len(args) != 2 {
		return nil, errors.New("нужно указать логин и пароль или использовать --interactive")
	}

	return &models.UsernamePassword{
		Username: args[0],
		Password: args[1],
	}, nil
}

// validateAuthRequest проверяет, что в структуре UsernamePassword заполнены логин и пароль.
// Возвращает ошибку, если какие-либо данные отсутствуют.
func validateAuthRequest(secret *models.UsernamePassword) error {
	if secret == nil {
		return errors.New("данные для аутентификации не заданы")
	}
	if secret.Username == "" {
		return errors.New("логин не может быть пустым")
	}
	if secret.Password == "" {
		return errors.New("пароль не может быть пустым")
	}
	return nil
}

// runAuth выполняет аутентификацию пользователя через HTTP или gRPC в зависимости от конфигурации клиента.
// Возвращает полученный токен или ошибку.
func runAuth(
	ctx context.Context,
	config *configs.ClientConfig,
	secret *models.UsernamePassword,
) (string, error) {
	var (
		token string
		err   error
	)

	switch {
	case config.HTTPClient != nil:
		token, err = services.AuthHTTP(ctx, config.HTTPClient, secret)
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к серверу по HTTP: %w", err)
		}
	case config.GRPCClient != nil:
		client := pb.NewAuthServiceClient(config.GRPCClient)
		token, err = services.AuthGRPC(ctx, client, secret)
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к серверу по gRPC: %w", err)
		}
	default:
		return "", errors.New("нет доступного клиента для подключения")
	}

	return token, nil
}

// setAuthEnv устанавливает переменные окружения с URL сервера и токеном для текущего процесса.
// Возвращает ошибку, если установка не удалась.
func setAuthEnv(serverURL, token string) error {
	if err := os.Setenv("GOPHKEEPER_SERVER_URL", serverURL); err != nil {
		return errors.New("не удалось установить переменную окружения для адреса сервера")
	}
	if err := os.Setenv("GOPHKEEPER_TOKEN", token); err != nil {
		return errors.New("не удалось установить переменную окружения для токена")
	}
	return nil
}
