package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RegisterListSecretCommand регистрирует команду CLI "list-secrets" для получения секретов пользователя.
// Параметры:
//
//	--server-url - URL сервера API (обязательный)
//	--token - JWT токен для авторизации (обязательный)
//	--secret-type - тип секрета (необязательный, если не указан — вернутся все секреты)
func RegisterListSecretCommand(root *cobra.Command) {
	var serverURL string
	var token string
	var secretType string

	cmd := &cobra.Command{
		Use:   "list-secrets",
		Short: "Получить все секреты пользователя",
		Long:  `Запрашивает с сервера все секреты указанного типа и выводит их.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverURL == "" {
				return fmt.Errorf("параметр --server-url обязателен")
			}
			if token == "" {
				return fmt.Errorf("параметр --token обязателен")
			}

			protocol, err := extractProtocol(serverURL)
			if err != nil {
				return fmt.Errorf("не удалось определить протокол из server-url: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			formatted, err := runListSecrets(ctx, protocol, serverURL, token, secretType)
			if err != nil {
				return err
			}

			fmt.Println(formatted)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URI сервера API (обязательный параметр)")
	cmd.Flags().StringVar(&token, "token", "", "JWT токен для авторизации (обязательный параметр)")
	cmd.Flags().StringVar(&secretType, "secret-type", "", "Тип секрета (необязательный параметр, если не указан — будут получены все секреты)")

	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("token")

	root.AddCommand(cmd)
}

// extractProtocol извлекает протокол из переданного URL.
// Поддерживаются grpc, http и https.
func extractProtocol(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(u.Scheme) {
	case "grpc":
		return "grpc", nil
	case "http", "https":
		return "http", nil
	default:
		return "", fmt.Errorf("неизвестный протокол: %s", u.Scheme)
	}
}

// runListSecrets вызывает нужную функцию для получения секретов в зависимости от типа секрета и протокола.
// Если secretType пустая строка — возвращает все секреты.
func runListSecrets(
	ctx context.Context,
	protocol string,
	serverURL string,
	token string,
	secretType string,
) (string, error) {
	switch secretType {
	case models.SecretTypeBankCard:
		return listBankCards(ctx, protocol, serverURL, token)
	case models.SecretTypeBinary:
		return listBinarySecrets(ctx, protocol, serverURL, token)
	case models.SecretTypeUsernamePassword:
		return listUsernamePasswordSecrets(ctx, protocol, serverURL, token)
	case models.SecretTypeText:
		return listTextSecrets(ctx, protocol, serverURL, token)
	case "":
		return listAllSecrets(ctx, protocol, serverURL, token)
	default:
		return "", fmt.Errorf("неизвестный тип секрета: %s", secretType)
	}
}

// listBankCards получает список секретов типа "bank-card" по протоколу http или grpc.
func listBankCards(ctx context.Context, protocol, serverURL, token string) (string, error) {
	if protocol == "http" {
		cfg, err := configs.NewClientConfig(configs.WithHTTPClient(serverURL))
		if err != nil {
			return "", fmt.Errorf("не удалось создать HTTP клиента: %w", err)
		}
		facade := facades.NewSecretBankCardListHTTPFacade(cfg.HTTPClient)
		secrets, err := facade.List(ctx, token)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return "", fmt.Errorf("ошибка форматирования ответа: %w", err)
		}
		return string(data), nil

	} else if protocol == "grpc" {
		conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к gRPC серверу: %w", err)
		}
		defer conn.Close()

		client := facades.NewSecretBankCardListGRPCFacade(pb.NewSecretBankCardServiceClient(conn))
		secrets, err := client.List(ctx, token)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return "", fmt.Errorf("ошибка форматирования ответа: %w", err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("неподдерживаемый протокол: %s", protocol)
}

// listBinarySecrets получает список секретов типа "binary" по протоколу http или grpc.
func listBinarySecrets(ctx context.Context, protocol, serverURL, token string) (string, error) {
	if protocol == "http" {
		cfg, err := configs.NewClientConfig(configs.WithHTTPClient(serverURL))
		if err != nil {
			return "", fmt.Errorf("не удалось создать HTTP клиента: %w", err)
		}
		facade := facades.NewSecretBankCardListHTTPFacade(cfg.HTTPClient)
		secrets, err := facade.List(ctx, token)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return "", fmt.Errorf("ошибка форматирования ответа: %w", err)
		}
		return string(data), nil

	} else if protocol == "grpc" {
		conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к gRPC серверу: %w", err)
		}
		defer conn.Close()

		client := facades.NewSecretBinaryListGRPCFacade(pb.NewSecretBinaryServiceClient(conn))
		secrets, err := client.List(ctx, token)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return "", fmt.Errorf("ошибка форматирования ответа: %w", err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("неподдерживаемый протокол: %s", protocol)
}

// listTextSecrets получает список секретов типа "text" по протоколу http или grpc.
func listTextSecrets(ctx context.Context, protocol, serverURL, token string) (string, error) {
	if protocol == "http" {
		cfg, err := configs.NewClientConfig(configs.WithHTTPClient(serverURL))
		if err != nil {
			return "", fmt.Errorf("не удалось создать HTTP клиента: %w", err)
		}
		facade := facades.NewTextListFacade(cfg.HTTPClient)
		secrets, err := facade.List(ctx, token)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return "", fmt.Errorf("ошибка форматирования ответа: %w", err)
		}
		return string(data), nil

	} else if protocol == "grpc" {
		conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к gRPC серверу: %w", err)
		}
		defer conn.Close()

		client := facades.NewTextListGRPCFacade(pb.NewSecretTextServiceClient(conn))
		secrets, err := client.List(ctx, token)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return "", fmt.Errorf("ошибка форматирования ответа: %w", err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("неподдерживаемый протокол: %s", protocol)
}

// listUsernamePasswordSecrets получает список секретов типа "username-password" по протоколу http или grpc.
func listUsernamePasswordSecrets(ctx context.Context, protocol, serverURL, token string) (string, error) {
	if protocol == "http" {
		cfg, err := configs.NewClientConfig(configs.WithHTTPClient(serverURL))
		if err != nil {
			return "", fmt.Errorf("не удалось создать HTTP клиента: %w", err)
		}
		facade := facades.NewSecretUsernamePasswordListHTTPFacade(cfg.HTTPClient)
		secrets, err := facade.List(ctx, token)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return "", fmt.Errorf("ошибка форматирования ответа: %w", err)
		}
		return string(data), nil

	} else if protocol == "grpc" {
		conn, err := grpc.Dial(serverURL, grpc.WithInsecure())
		if err != nil {
			return "", fmt.Errorf("не удалось подключиться к gRPC серверу: %w", err)
		}
		defer conn.Close()

		client := facades.NewSecretUsernamePasswordListGRPCFacade(pb.NewSecretUsernamePasswordServiceClient(conn))
		secrets, err := client.List(ctx, token)
		if err != nil {
			return "", err
		}
		data, err := json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return "", fmt.Errorf("ошибка форматирования ответа: %w", err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("неподдерживаемый протокол: %s", protocol)
}

// listAllSecrets получает все типы секретов и объединяет их в один JSON-массив.
// Поддерживает протоколы http и grpc.
func listAllSecrets(
	ctx context.Context,
	protocol, serverURL, token string,
) (string, error) {
	secretTypes := []string{
		models.SecretTypeUsernamePassword,
		models.SecretTypeText,
		models.SecretTypeBinary,
		models.SecretTypeBankCard,
	}

	allSecrets := []map[string]any{}

	for _, t := range secretTypes {
		var formatted string
		var err error

		switch t {
		case models.SecretTypeUsernamePassword:
			formatted, err = listUsernamePasswordSecrets(ctx, protocol, serverURL, token)
		case models.SecretTypeText:
			formatted, err = listTextSecrets(ctx, protocol, serverURL, token)
		case models.SecretTypeBinary:
			formatted, err = listBinarySecrets(ctx, protocol, serverURL, token)
		case models.SecretTypeBankCard:
			formatted, err = listBankCards(ctx, protocol, serverURL, token)
		}

		if err != nil {
			return "", fmt.Errorf("ошибка при получении секретов типа %s: %w", t, err)
		}

		var secrets []map[string]any
		if err := json.Unmarshal([]byte(formatted), &secrets); err != nil {
			return "", fmt.Errorf("не удалось распарсить секреты типа %s: %w", t, err)
		}
		allSecrets = append(allSecrets, secrets...)
	}

	data, err := json.MarshalIndent(allSecrets, "", "  ")
	if err != nil {
		return "", fmt.Errorf("ошибка обработки данных: %w", err)
	}

	return string(data), nil
}
