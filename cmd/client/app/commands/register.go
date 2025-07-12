package commands

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// Register выполняет регистрацию пользователя.
// Принимает контекст ctx, аргументы args, флаги flags, переменные окружения envs и reader для интерактивного ввода.
// Возвращает ошибку в случае неудачи.
func Register(
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

	serverURL, interactive, err := parseRegisterFlags(flags)
	if err != nil {
		return err
	}

	if interactive {
		secret, err = parseRegisterFlagsInteractive(reader)
	} else {
		secret, err = parseRegisterArgs(args)
	}
	if err != nil {
		return err
	}

	if err = validateRegisterRequest(secret); err != nil {
		return err
	}

	config, err := newRegisterConfig(serverURL)
	if err != nil {
		return err
	}

	token, err := runRegister(ctx, config, secret)
	if err != nil {
		return err
	}

	if err = setRegisterEnv(serverURL, token); err != nil {
		return err
	}

	return nil
}

// parseRegisterFlags извлекает значения serverURL и interactive из переданных флагов.
// Возвращает serverURL (строка), interactive (bool) и ошибку, если значение interactive некорректно.
func parseRegisterFlags(flags map[string]string) (string, bool, error) {
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

// parseRegisterFlagsInteractive запрашивает у пользователя логин и пароль из stdin или другого io.Reader.
// Возвращает структуру UsernamePassword и ошибку, если ввод не удался.
func parseRegisterFlagsInteractive(r io.Reader) (*models.UsernamePassword, error) {
	reader := bufio.NewReader(r)

	print("Введите логин: ")
	inputLogin, err := reader.ReadString('\n')
	if err != nil {
		return nil, errors.New("ошибка при вводе логина")
	}
	login := strings.TrimSpace(inputLogin)

	print("Введите пароль: ")
	inputPassword, err := reader.ReadString('\n')
	if err != nil {
		return nil, errors.New("ошибка при вводе пароля")
	}
	password := strings.TrimSpace(inputPassword)

	return &models.UsernamePassword{
		Username: login,
		Password: password,
	}, nil
}

// parseRegisterArgs проверяет аргументы командной строки и возвращает структуру UsernamePassword.
// Если передано неверное количество аргументов, возвращает ошибку.
func parseRegisterArgs(args []string) (*models.UsernamePassword, error) {
	if len(args) != 2 {
		return nil, errors.New("нужно указать логин и пароль или использовать --interactive")
	}

	return &models.UsernamePassword{
		Username: args[0],
		Password: args[1],
	}, nil
}

// validateRegisterRequest проверяет, что логин и пароль не пустые.
// Возвращает ошибку, если какие-либо данные отсутствуют.
func validateRegisterRequest(secret *models.UsernamePassword) error {
	if secret == nil {
		return errors.New("данные для регистрации не заданы")
	}
	if secret.Username == "" {
		return errors.New("логин не может быть пустым")
	}
	if secret.Password == "" {
		return errors.New("пароль не может быть пустым")
	}
	return nil
}

// newRegisterConfig создает конфигурацию клиента для регистрации на основе serverURL.
// Поддерживает HTTP и gRPC протоколы.
// Возвращает указатель на ClientConfig или ошибку при неверном формате URL или проблемах с созданием.
func newRegisterConfig(serverURL string) (*configs.ClientConfig, error) {
	var opts []configs.ClientConfigOpt

	switch {
	case strings.HasPrefix(serverURL, "http://"), strings.HasPrefix(serverURL, "https://"):
		opts = append(opts, configs.WithHTTPClient(serverURL))
	case strings.HasPrefix(serverURL, "grpc://"):
		opts = append(opts, configs.WithGRPCClient(serverURL))
	default:
		return nil, errors.New("неверный формат URL сервера")
	}

	config, err := configs.NewClientConfig(opts...)
	if err != nil {
		return nil, errors.New("не удалось создать конфигурацию клиента")
	}

	return config, nil
}

// runRegister выполняет регистрацию пользователя через HTTP или gRPC в зависимости от конфигурации клиента.
// Возвращает полученный токен или ошибку.
func runRegister(
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
		token, err = services.RegisterHTTP(ctx, config.HTTPClient, secret)
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к серверу по HTTP: %w", err)
		}
	case config.GRPCClient != nil:
		client := pb.NewRegisterServiceClient(config.GRPCClient)
		token, err = services.RegisterGRPC(ctx, client, secret)
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к серверу по gRPC: %w", err)
		}
	default:
		return "", errors.New("нет доступного клиента для подключения")
	}

	return token, nil
}

// setRegisterEnv устанавливает переменные окружения для адреса сервера и токена регистрации.
// Возвращает ошибку, если не удалось установить переменные окружения.
func setRegisterEnv(serverURL, token string) error {
	if err := os.Setenv("GOPHKEEPER_SERVER_URL", serverURL); err != nil {
		return errors.New("не удалось установить переменную окружения для адреса сервера")
	}
	if err := os.Setenv("GOPHKEEPER_TOKEN", token); err != nil {
		return errors.New("не удалось установить переменную окружения для токена")
	}
	return nil
}
