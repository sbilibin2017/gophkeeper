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

// Login выполняет вход пользователя в систему.
func Login(
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

	serverURL, interactive, err := parseLoginFlags(flags)
	if err != nil {
		return err
	}

	if interactive {
		secret, err = parseLoginFlagsInteractive(reader)
	} else {
		secret, err = parseLoginArgs(args)
	}
	if err != nil {
		return err
	}

	if err = validateLoginRequest(secret); err != nil {
		return err
	}

	config, err := newLoginConfig(serverURL)
	if err != nil {
		return err
	}

	token, err := runLogin(ctx, config, secret)
	if err != nil {
		return err
	}

	if err = setLoginEnv(serverURL, token); err != nil {
		return err
	}

	fmt.Println("Вход выполнен. Токен сохранён в переменной окружения GOPHKEEPER_TOKEN.")
	return nil
}

// parseLoginFlags извлекает serverURL и interactive из флагов.
func parseLoginFlags(flags map[string]string) (string, bool, error) {
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

// parseLoginFlagsInteractive запрашивает логин и пароль у пользователя.
func parseLoginFlagsInteractive(r io.Reader) (*models.UsernamePassword, error) {
	reader := bufio.NewReader(r)

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

	return &models.UsernamePassword{
		Username: login,
		Password: password,
	}, nil
}

// parseLoginArgs парсит логин и пароль из аргументов.
func parseLoginArgs(args []string) (*models.UsernamePassword, error) {
	if len(args) != 2 {
		return nil, errors.New("нужно указать логин и пароль или использовать --interactive")
	}

	return &models.UsernamePassword{
		Username: args[0],
		Password: args[1],
	}, nil
}

// validateLoginRequest проверяет, что логин и пароль указаны.
func validateLoginRequest(secret *models.UsernamePassword) error {
	if secret == nil {
		return errors.New("данные для входа не заданы")
	}
	if secret.Username == "" {
		return errors.New("логин не может быть пустым")
	}
	if secret.Password == "" {
		return errors.New("пароль не может быть пустым")
	}
	return nil
}

// newLoginConfig создает клиентскую конфигурацию.
func newLoginConfig(serverURL string) (*configs.ClientConfig, error) {
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

// runLogin выполняет авторизацию через HTTP или gRPC.
func runLogin(
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
		token, err = services.LoginHTTP(ctx, config.HTTPClient, secret)
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к серверу по HTTP: %w", err)
		}
	case config.GRPCClient != nil:
		client := pb.NewLoginServiceClient(config.GRPCClient)
		token, err = services.LoginGRPC(ctx, client, secret)
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к серверу по gRPC: %w", err)
		}
	default:
		return "", errors.New("нет доступного клиента для подключения")
	}

	return token, nil
}

// setLoginEnv сохраняет токен и адрес сервера в переменные окружения.
func setLoginEnv(serverURL, token string) error {
	if err := os.Setenv("GOPHKEEPER_SERVER_URL", serverURL); err != nil {
		return errors.New("не удалось установить переменную окружения для адреса сервера")
	}
	if err := os.Setenv("GOPHKEEPER_TOKEN", token); err != nil {
		return errors.New("не удалось установить переменную окружения для токена")
	}
	return nil
}
