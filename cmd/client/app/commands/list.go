package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/protocol"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func RegisterListSecretCommand(root *cobra.Command) {
	var serverURL string
	var token string
	var secretType string

	cmd := &cobra.Command{
		Use:   "list-secrets",
		Short: "Получить все секреты пользователя",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := validateListInput(serverURL, token)
			if err != nil {
				return err
			}

			proto := protocol.GetProtocolFromURL(serverURL)
			if proto == "" {
				return fmt.Errorf("не удалось определить протокол по server-url: %s", serverURL)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var result string

			switch proto {
			case protocol.HTTP, protocol.HTTPS:
				cfg, err := configs.NewClientConfig(configs.WithClientConfigHTTPClient(serverURL))
				if err != nil {
					return fmt.Errorf("ошибка создания HTTP клиента: %w", err)
				}
				result, err = listSecretsHTTP(ctx, cfg.HTTPClient, token, secretType)

			case protocol.GRPC:
				addr := strings.TrimPrefix(serverURL, "grpc://")
				cfg, err := configs.NewClientConfig(configs.WithClientConfigGRPCClient(addr))
				if err != nil {
					return fmt.Errorf("ошибка создания gRPC клиента: %w", err)
				}
				result, err = listSecretsGRPC(ctx, cfg.GRPCClient, token, secretType)

			default:
				return fmt.Errorf("неподдерживаемый протокол: %s", proto)
			}

			if err != nil {
				return err
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URL сервера API (обязательный)")
	cmd.Flags().StringVar(&token, "token", "", "JWT токен для авторизации (обязательный)")
	cmd.Flags().StringVar(&secretType, "secret-type", "", "Тип секрета (опционально)")

	root.AddCommand(cmd)
}

func validateListInput(serverURL, token string) error {
	if serverURL == "" {
		return fmt.Errorf("--server-url обязателен")
	}
	if token == "" {
		return fmt.Errorf("--token обязателен")
	}
	return nil
}

// Вызовы через HTTP
func listSecretsHTTP(ctx context.Context, c *resty.Client, token, secretType string) (string, error) {
	switch secretType {
	case models.SecretTypeBankCard:
		secrets, err := client.ListSecretBankCardHTTP(ctx, c, token)
		if err != nil {
			return "", err
		}
		return marshalJSON(secrets)

	case models.SecretTypeBinary:
		secrets, err := client.ListSecretBinaryHTTP(ctx, c, token)
		if err != nil {
			return "", err
		}
		return marshalJSON(secrets)

	case models.SecretTypeUsernamePassword:
		secrets, err := client.ListSecretUsernamePasswordHTTP(ctx, c, token)
		if err != nil {
			return "", err
		}
		return marshalJSON(secrets)

	case models.SecretTypeText:
		secrets, err := client.ListSecretTextHTTP(ctx, c, token)
		if err != nil {
			return "", err
		}
		return marshalJSON(secrets)

	case "":
		return listAllSecretsHTTP(ctx, c, token)

	default:
		return "", fmt.Errorf("неизвестный тип секрета: %s", secretType)
	}
}

func listAllSecretsHTTP(ctx context.Context, c *resty.Client, token string) (string, error) {
	allSecrets := []any{}

	secretTypes := []string{
		models.SecretTypeUsernamePassword,
		models.SecretTypeText,
		models.SecretTypeBinary,
		models.SecretTypeBankCard,
	}

	for _, t := range secretTypes {
		var (
			secretsJSON string
			err         error
		)

		switch t {
		case models.SecretTypeUsernamePassword:
			secretsJSON, err = listSecretsHTTP(ctx, c, token, t)
		case models.SecretTypeText:
			secretsJSON, err = listSecretsHTTP(ctx, c, token, t)
		case models.SecretTypeBinary:
			secretsJSON, err = listSecretsHTTP(ctx, c, token, t)
		case models.SecretTypeBankCard:
			secretsJSON, err = listSecretsHTTP(ctx, c, token, t)
		}

		if err != nil {
			return "", fmt.Errorf("ошибка получения секретов типа %s: %w", t, err)
		}

		var items []any
		if err := json.Unmarshal([]byte(secretsJSON), &items); err != nil {
			return "", fmt.Errorf("ошибка парсинга секретов типа %s: %w", t, err)
		}

		allSecrets = append(allSecrets, items...)
	}

	return marshalJSON(allSecrets)
}

// Вызовы через gRPC
func listSecretsGRPC(ctx context.Context, conn *grpc.ClientConn, token, secretType string) (string, error) {
	switch secretType {
	case models.SecretTypeBankCard:
		c := pb.NewSecretBankCardServiceClient(conn)
		secrets, err := client.ListSecretBankCardGRPC(ctx, c, token)
		if err != nil {
			return "", err
		}
		return marshalJSON(secrets)

	case models.SecretTypeBinary:
		c := pb.NewSecretBinaryServiceClient(conn)
		secrets, err := client.ListSecretBinaryGRPC(ctx, c, token)
		if err != nil {
			return "", err
		}
		return marshalJSON(secrets)

	case models.SecretTypeUsernamePassword:
		c := pb.NewSecretUsernamePasswordServiceClient(conn)
		secrets, err := client.ListSecretUsernamePasswordGRPC(ctx, c, token)
		if err != nil {
			return "", err
		}
		return marshalJSON(secrets)

	case models.SecretTypeText:
		c := pb.NewSecretTextServiceClient(conn)
		secrets, err := client.ListSecretTextGRPC(ctx, c, token)
		if err != nil {
			return "", err
		}
		return marshalJSON(secrets)

	case "":
		return listAllSecretsGRPC(ctx, conn, token)

	default:
		return "", fmt.Errorf("неизвестный тип секрета: %s", secretType)
	}
}

func listAllSecretsGRPC(ctx context.Context, conn *grpc.ClientConn, token string) (string, error) {
	allSecrets := []any{}

	secretTypes := []string{
		models.SecretTypeUsernamePassword,
		models.SecretTypeText,
		models.SecretTypeBinary,
		models.SecretTypeBankCard,
	}

	for _, t := range secretTypes {
		var (
			result string
			err    error
		)

		switch t {
		case models.SecretTypeUsernamePassword:
			result, err = listSecretsGRPC(ctx, conn, token, t)
		case models.SecretTypeText:
			result, err = listSecretsGRPC(ctx, conn, token, t)
		case models.SecretTypeBinary:
			result, err = listSecretsGRPC(ctx, conn, token, t)
		case models.SecretTypeBankCard:
			result, err = listSecretsGRPC(ctx, conn, token, t)
		}

		if err != nil {
			return "", fmt.Errorf("ошибка получения секретов типа %s: %w", t, err)
		}

		var items []any
		if err := json.Unmarshal([]byte(result), &items); err != nil {
			return "", fmt.Errorf("ошибка парсинга секретов типа %s: %w", t, err)
		}

		allSecrets = append(allSecrets, items...)
	}

	return marshalJSON(allSecrets)
}

func marshalJSON(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("ошибка форматирования JSON: %w", err)
	}
	return string(data), nil
}
