package commands

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/jmoiron/sqlx"

	"github.com/sbilibin2017/gophkeeper/internal/client"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/configs/protocol"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func RegisterSyncCommand(root *cobra.Command) {
	var serverURL string
	var token string
	var resolve string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Синхронизация клиента с сервером",
		Long: `Синхронизирует локальные данные клиента с сервером.
Допускается выбор стратегии разрешения конфликтов:
- client: клиентские данные имеют приоритет
- server: серверные данные имеют приоритет
- interactive: выбор вручную`,
		Example: `  gophkeeper sync --server-url=http://localhost:8080 --resolve=client`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Проверка стратегии разрешения конфликтов
			switch resolve {
			case "client", "server", "interactive":
			default:
				return fmt.Errorf("неподдерживаемая стратегия resolve: %s (возможные: client, server, interactive)", resolve)
			}

			if serverURL == "" {
				return fmt.Errorf("не указан адрес сервера (--server-url)")
			}
			if token == "" {
				return fmt.Errorf("не указан JWT токен (--token)")
			}

			ctx := context.Background()

			// Создаем конфигурацию клиента в зависимости от протокола сервера
			cfg, err := func() (*configs.ClientConfig, error) {
				const dbName = "gophkeeper_client.db"

				proto := protocol.GetProtocolFromURL(serverURL)

				switch proto {
				case protocol.HTTP, protocol.HTTPS:
					return configs.NewClientConfig(
						configs.WithClientConfigHTTPClient(serverURL),
						configs.WithClientConfigDB(dbName),
					)
				case protocol.GRPC:
					return configs.NewClientConfig(
						configs.WithClientConfigGRPCClient(serverURL),
						configs.WithClientConfigDB(dbName),
					)
				default:
					return nil, fmt.Errorf("неподдерживаемый протокол сервера: %s", serverURL)
				}
			}()
			if err != nil {
				return fmt.Errorf("не удалось создать конфиг клиента: %w", err)
			}

			// Получаем локальные секреты из базы
			textSecrets, err := client.GetAllSecretsTextRequest(ctx, cfg.DB)
			if err != nil {
				return fmt.Errorf("ошибка получения текстовых секретов: %w", err)
			}

			binarySecrets, err := client.GetAllSecretsBinaryRequest(ctx, cfg.DB)
			if err != nil {
				return fmt.Errorf("ошибка получения бинарных секретов: %w", err)
			}

			bankCardSecrets, err := client.GetAllSecretsBankCardRequest(ctx, cfg.DB)
			if err != nil {
				return fmt.Errorf("ошибка получения банковских карт: %w", err)
			}

			loginPasswordSecrets, err := client.GetAllSecretsUsernamePasswordRequest(ctx, cfg.DB)
			if err != nil {
				return fmt.Errorf("ошибка получения логинов и паролей: %w", err)
			}

			proto := protocol.GetProtocolFromURL(serverURL)

			// В зависимости от протокола вызываем соответствующие методы синхронизации
			switch proto {
			case protocol.HTTP, protocol.HTTPS:
				client := cfg.HTTPClient

				if err := syncTextHTTP(ctx, client, cfg.DB, token, resolve, textSecrets); err != nil {
					return err
				}
				if err := syncBinaryHTTP(ctx, client, cfg.DB, token, resolve, binarySecrets); err != nil {
					return err
				}
				if err := syncBankCardHTTP(ctx, client, cfg.DB, token, resolve, bankCardSecrets); err != nil {
					return err
				}
				if err := syncUsernamePasswordHTTP(ctx, client, cfg.DB, token, resolve, loginPasswordSecrets); err != nil {
					return err
				}

			case protocol.GRPC:
				client := cfg.GRPCClient

				if err := syncTextGRPC(ctx, client, cfg.DB, token, resolve, textSecrets); err != nil {
					return err
				}
				if err := syncBinaryGRPC(ctx, client, cfg.DB, token, resolve, binarySecrets); err != nil {
					return err
				}
				if err := syncBankCardGRPC(ctx, client, cfg.DB, token, resolve, bankCardSecrets); err != nil {
					return err
				}
				if err := syncUsernamePasswordGRPC(ctx, client, cfg.DB, token, resolve, loginPasswordSecrets); err != nil {
					return err
				}

			default:
				return fmt.Errorf("неподдерживаемый протокол сервера: %s", serverURL)
			}

			fmt.Println("Синхронизация завершена успешно.")
			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "", "URI сервера синхронизации (обязательный)")
	cmd.Flags().StringVar(&resolve, "resolve", "", "Стратегия разрешения конфликтов: client, server, interactive")
	cmd.Flags().StringVar(&token, "token", "", "JWT токен для авторизации (обязательный параметр)")

	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("server-url")
	_ = cmd.MarkFlagRequired("resolve")

	root.AddCommand(cmd)
}

func syncTextHTTP(
	ctx context.Context,
	c *resty.Client,
	db *sqlx.DB,
	token string,
	resolve string,
	localSecrets []models.SecretTextGetRequest,
) error {
	makeSaveReq := func(secret *models.SecretTextGetResponse) models.SecretTextSaveRequest {
		return models.SecretTextSaveRequest{
			SecretName: secret.SecretName,
			Content:    secret.Content,
			Meta:       secret.Meta,
		}
	}

	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			fullSecret, err := client.GetSecretTextByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный полный текстовый секрет %s: %v\n", secret.SecretName, err)
				continue
			}
			saveReq := makeSaveReq(fullSecret)
			if err := client.SaveSecretTextHTTP(ctx, c, token, saveReq); err != nil {
				fmt.Printf("Ошибка при отправке текстового секрета %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		// Пусто — логика "загрузить всё с сервера" может быть реализована отдельно
		return nil

	case "interactive":
		for _, secret := range localSecrets {
			localSecret, err := client.GetSecretTextByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный секрет %s: %v\n", secret.SecretName, err)
				continue
			}

			serverSecret, err := client.GetSecretTextHTTP(ctx, c, token, secret.SecretName)
			if err != nil {
				// Нет серверной версии — отправляем локальную
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretTextHTTP(ctx, c, token, saveReq); err != nil {
					fmt.Printf("Ошибка при отправке локального секрета %s на сервер: %v\n", secret.SecretName, err)
				}
				continue
			}

			// Всегда предлагать выбор пользователю
			fmt.Printf("Конфликт секрета %s:\n", secret.SecretName)
			fmt.Println("Локальная версия:")
			fmt.Println(localSecret.Content)
			fmt.Println("Серверная версия:")
			fmt.Println(serverSecret.Content)
			fmt.Print("Выберите версию для сохранения (client/server): ")

			var choice string
			_, err = fmt.Scanln(&choice)
			if err != nil {
				fmt.Printf("Ошибка при вводе выбора: %v\n", err)
				continue
			}

			switch choice {
			case "client":
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretTextHTTP(ctx, c, token, saveReq); err != nil {
					fmt.Printf("Ошибка при сохранении локального секрета %s на сервере: %v\n", secret.SecretName, err)
				}
			case "server":
				continue
			default:
				fmt.Println("Неверный выбор. Пропуск секрета.")
			}
		}
	}
	return nil
}

func syncTextGRPC(
	ctx context.Context,
	c grpc.ClientConnInterface,
	db *sqlx.DB,
	token string,
	resolve string,
	localSecrets []models.SecretTextGetRequest,
) error {
	cl := pb.NewSecretTextServiceClient(c)

	makeSaveReq := func(secret *models.SecretTextGetResponse) models.SecretTextSaveRequest {
		return models.SecretTextSaveRequest{
			SecretName: secret.SecretName,
			Content:    secret.Content,
			Meta:       secret.Meta,
		}
	}

	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			fullSecret, err := client.GetSecretTextByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный полный текстовый секрет %s: %v\n", secret.SecretName, err)
				continue
			}
			saveReq := makeSaveReq(fullSecret)
			if err := client.SaveSecretTextGRPC(ctx, cl, token, saveReq); err != nil {
				fmt.Printf("Ошибка при отправке текстового секрета %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		// Пусто — логика "загрузить всё с сервера" может быть реализована отдельно
		return nil

	case "interactive":
		for _, secret := range localSecrets {
			localSecret, err := client.GetSecretTextByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный секрет %s: %v\n", secret.SecretName, err)
				continue
			}

			serverSecret, err := client.GetSecretTextGRPC(ctx, cl, token, secret.SecretName)
			if err != nil {
				// Нет серверной версии — отправляем локальную
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretTextGRPC(ctx, cl, token, saveReq); err != nil {
					fmt.Printf("Ошибка при отправке локального секрета %s на сервер: %v\n", secret.SecretName, err)
				}
				continue
			}

			// Всегда предлагать выбор пользователю
			fmt.Printf("Конфликт секрета %s:\n", secret.SecretName)
			fmt.Println("Локальная версия:")
			fmt.Println(localSecret.Content)
			fmt.Println("Серверная версия:")
			fmt.Println(serverSecret.Content)
			fmt.Print("Выберите версию для сохранения (client/server): ")

			var choice string
			_, err = fmt.Scanln(&choice)
			if err != nil {
				fmt.Printf("Ошибка при вводе выбора: %v\n", err)
				continue
			}

			switch choice {
			case "client":
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretTextGRPC(ctx, cl, token, saveReq); err != nil {
					fmt.Printf("Ошибка при сохранении локального секрета %s на сервере: %v\n", secret.SecretName, err)
				}
			case "server":
				continue
			default:
				fmt.Println("Неверный выбор. Пропуск секрета.")
			}
		}
	}
	return nil
}

func syncBinaryHTTP(
	ctx context.Context,
	c *resty.Client,
	db *sqlx.DB,
	token string,
	resolve string,
	localSecrets []models.SecretBinaryGetRequest,
) error {
	makeSaveReq := func(secret *models.SecretBinaryGetResponse) models.SecretBinarySaveRequest {
		return models.SecretBinarySaveRequest{
			SecretName: secret.SecretName,
			Data:       secret.Data,
			Meta:       secret.Meta,
		}
	}

	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			fullSecret, err := client.GetSecretBinaryByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный полный бинарный секрет %s: %v\n", secret.SecretName, err)
				continue
			}
			saveReq := makeSaveReq(fullSecret)
			if err := client.SaveSecretBinaryHTTP(ctx, c, token, saveReq); err != nil {
				fmt.Printf("Ошибка при отправке бинарного секрета %s: %v\n", secret.SecretName, err)
			}
		}

	case "server":
		// Пусто — логика "загрузить всё с сервера" может быть реализована отдельно
		return nil

	case "interactive":
		for _, secret := range localSecrets {
			localSecret, err := client.GetSecretBinaryByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный бинарный секрет %s: %v\n", secret.SecretName, err)
				continue
			}

			serverSecret, err := client.GetSecretBinaryHTTP(ctx, c, token, secret.SecretName)
			if err != nil {
				fmt.Printf("Серверный бинарный секрет %s не найден.\n", secret.SecretName)
				fmt.Printf("Локальный бинарный секрет %s доступен.\n", secret.SecretName)
				fmt.Print("Отправить локальную версию на сервер? (yes/no): ")

				var choice string
				_, err := fmt.Scanln(&choice)
				if err != nil {
					fmt.Printf("Ошибка при вводе выбора: %v\n", err)
					continue
				}

				if choice == "yes" {
					saveReq := makeSaveReq(localSecret)
					if err := client.SaveSecretBinaryHTTP(ctx, c, token, saveReq); err != nil {
						fmt.Printf("Ошибка при отправке бинарного секрета %s: %v\n", secret.SecretName, err)
					}
				}
				continue
			}

			fmt.Printf("Конфликт бинарного секрета %s:\n", secret.SecretName)
			fmt.Println("Локальная версия (размер в байтах):", len(localSecret.Data))
			fmt.Println("Серверная версия (размер в байтах):", len(serverSecret.Data))
			fmt.Print("Выберите версию для сохранения (client/server/skip): ")

			var choice string
			_, err = fmt.Scanln(&choice)
			if err != nil {
				fmt.Printf("Ошибка при вводе выбора: %v\n", err)
				continue
			}

			switch choice {
			case "client":
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretBinaryHTTP(ctx, c, token, saveReq); err != nil {
					fmt.Printf("Ошибка при отправке локального бинарного секрета %s: %v\n", secret.SecretName, err)
				}
			case "server":
				continue // пропускаем, серверная версия считается актуальной
			case "skip":
				fmt.Println("Пропуск секрета.")
			default:
				fmt.Println("Неверный выбор. Пропуск секрета.")
			}
		}
	}
	return nil
}

func syncBinaryGRPC(
	ctx context.Context,
	grpcConn grpc.ClientConnInterface,
	db *sqlx.DB,
	token string,
	resolve string,
	localSecrets []models.SecretBinaryGetRequest,
) error {
	cl := pb.NewSecretBinaryServiceClient(grpcConn)

	makeSaveReq := func(secret *models.SecretBinaryGetResponse) models.SecretBinarySaveRequest {
		return models.SecretBinarySaveRequest{
			SecretName: secret.SecretName,
			Data:       secret.Data,
			Meta:       secret.Meta,
		}
	}

	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			fullSecret, err := client.GetSecretBinaryByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный полный бинарный секрет %s: %v\n", secret.SecretName, err)
				continue
			}
			saveReq := makeSaveReq(fullSecret)
			if err := client.SaveSecretBinaryGRPC(ctx, cl, token, saveReq); err != nil {
				fmt.Printf("Ошибка при отправке бинарного секрета %s: %v\n", secret.SecretName, err)
			}
		}

	case "server":
		// Пусто — логика "загрузить всё с сервера" может быть реализована отдельно
		return nil

	case "interactive":
		for _, secret := range localSecrets {
			localSecret, err := client.GetSecretBinaryByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный бинарный секрет %s: %v\n", secret.SecretName, err)
				continue
			}

			serverSecret, err := client.GetSecretBinaryGRPC(ctx, cl, token, secret.SecretName)
			if err != nil {
				fmt.Printf("Серверный бинарный секрет %s не найден.\n", secret.SecretName)
				fmt.Printf("Локальный бинарный секрет %s доступен.\n", secret.SecretName)
				fmt.Print("Отправить локальную версию на сервер? (yes/no): ")

				var choice string
				_, err := fmt.Scanln(&choice)
				if err != nil {
					fmt.Printf("Ошибка при вводе выбора: %v\n", err)
					continue
				}

				if choice == "yes" {
					saveReq := makeSaveReq(localSecret)
					if err := client.SaveSecretBinaryGRPC(ctx, cl, token, saveReq); err != nil {
						fmt.Printf("Ошибка при отправке бинарного секрета %s: %v\n", secret.SecretName, err)
					}
				}
				continue
			}

			fmt.Printf("Конфликт бинарного секрета %s:\n", secret.SecretName)
			fmt.Println("Локальная версия (размер в байтах):", len(localSecret.Data))
			fmt.Println("Серверная версия (размер в байтах):", len(serverSecret.Data))
			fmt.Print("Выберите версию для сохранения (client/server/skip): ")

			var choice string
			_, err = fmt.Scanln(&choice)
			if err != nil {
				fmt.Printf("Ошибка при вводе выбора: %v\n", err)
				continue
			}

			switch choice {
			case "client":
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretBinaryGRPC(ctx, cl, token, saveReq); err != nil {
					fmt.Printf("Ошибка при отправке локального бинарного секрета %s: %v\n", secret.SecretName, err)
				}
			case "server":
				continue // серверная версия считается актуальной
			case "skip":
				fmt.Println("Пропуск секрета.")
			default:
				fmt.Println("Неверный выбор. Пропуск секрета.")
			}
		}
	}
	return nil
}

func syncUsernamePasswordHTTP(
	ctx context.Context,
	c *resty.Client,
	db *sqlx.DB,
	token string,
	resolve string,
	localSecrets []models.SecretUsernamePasswordGetRequest,
) error {
	makeSaveReq := func(secret *models.SecretUsernamePasswordGetResponse) models.SecretUsernamePasswordSaveRequest {
		return models.SecretUsernamePasswordSaveRequest{
			SecretName: secret.SecretName,
			Username:   secret.Username,
			Password:   secret.Password,
			Meta:       secret.Meta,
		}
	}

	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			fullSecret, err := client.GetSecretUsernamePasswordByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный полный секрет логина/пароля %s: %v\n", secret.SecretName, err)
				continue
			}
			saveReq := makeSaveReq(fullSecret)
			if err := client.SaveSecretUsernamePasswordHTTP(ctx, c, token, saveReq); err != nil {
				fmt.Printf("Ошибка при отправке секрета логина/пароля %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		// Пусто — логика загрузки с сервера может быть реализована отдельно
		return nil
	case "interactive":
		for _, secret := range localSecrets {
			localSecret, err := client.GetSecretUsernamePasswordByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный секрет логина/пароля %s: %v\n", secret.SecretName, err)
				continue
			}

			serverSecret, err := client.GetSecretUsernamePasswordHTTP(ctx, c, token, secret.SecretName)
			if err != nil {
				fmt.Printf("Серверный секрет логина/пароля %s не найден.\n", secret.SecretName)
				fmt.Print("Отправить локальную версию на сервер? (yes/no): ")

				var choice string
				_, err := fmt.Scanln(&choice)
				if err != nil {
					fmt.Printf("Ошибка при вводе выбора: %v\n", err)
					continue
				}
				if choice == "yes" {
					saveReq := makeSaveReq(localSecret)
					if err := client.SaveSecretUsernamePasswordHTTP(ctx, c, token, saveReq); err != nil {
						fmt.Printf("Ошибка при отправке секрета логина/пароля %s: %v\n", localSecret.SecretName, err)
					}
				}
				continue
			}

			fmt.Printf("Конфликт секрета логина/пароля %s:\n", secret.SecretName)
			fmt.Println("Локальная версия:")
			fmt.Printf("Username: %s\nPassword: %s\n", localSecret.Username, localSecret.Password)
			fmt.Println("Серверная версия:")
			fmt.Printf("Username: %s\nPassword: %s\n", serverSecret.Username, serverSecret.Password)
			fmt.Print("Выберите версию для сохранения (client/server/skip): ")

			var choice string
			_, err = fmt.Scanln(&choice)
			if err != nil {
				fmt.Printf("Ошибка при вводе выбора: %v\n", err)
				continue
			}

			switch choice {
			case "client":
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretUsernamePasswordHTTP(ctx, c, token, saveReq); err != nil {
					fmt.Printf("Ошибка при сохранении локального секрета логина/пароля %s на сервере: %v\n", secret.SecretName, err)
				}
			case "server":
				continue
			case "skip":
				fmt.Println("Пропуск секрета.")
			default:
				fmt.Println("Неверный выбор. Пропуск секрета.")
			}
		}
	}
	return nil
}

func syncUsernamePasswordGRPC(
	ctx context.Context,
	grpcClient grpc.ClientConnInterface,
	db *sqlx.DB,
	token string,
	resolve string,
	localSecrets []models.SecretUsernamePasswordGetRequest,
) error {
	cl := pb.NewSecretUsernamePasswordServiceClient(grpcClient)

	makeSaveReq := func(secret *models.SecretUsernamePasswordGetResponse) models.SecretUsernamePasswordSaveRequest {
		return models.SecretUsernamePasswordSaveRequest{
			SecretName: secret.SecretName,
			Username:   secret.Username,
			Password:   secret.Password,
			Meta:       secret.Meta,
		}
	}

	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			fullSecret, err := client.GetSecretUsernamePasswordByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный полный секрет логина/пароля %s: %v\n", secret.SecretName, err)
				continue
			}
			saveReq := makeSaveReq(fullSecret)
			if err := client.SaveSecretUsernamePasswordGRPC(ctx, cl, token, saveReq); err != nil {
				fmt.Printf("Ошибка при отправке секрета логина/пароля %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		return nil
	case "interactive":
		for _, secret := range localSecrets {
			localSecret, err := client.GetSecretUsernamePasswordByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный секрет логина/пароля %s: %v\n", secret.SecretName, err)
				continue
			}

			serverSecret, err := client.GetSecretUsernamePasswordGRPC(ctx, cl, token, secret.SecretName)
			if err != nil {
				fmt.Printf("Серверный секрет логина/пароля %s не найден.\n", secret.SecretName)
				fmt.Print("Отправить локальную версию на сервер? (yes/no): ")

				var choice string
				_, err := fmt.Scanln(&choice)
				if err != nil {
					fmt.Printf("Ошибка при вводе выбора: %v\n", err)
					continue
				}
				if choice == "yes" {
					saveReq := makeSaveReq(localSecret)
					if err := client.SaveSecretUsernamePasswordGRPC(ctx, cl, token, saveReq); err != nil {
						fmt.Printf("Ошибка при отправке секрета логина/пароля %s: %v\n", localSecret.SecretName, err)
					}
				}
				continue
			}

			fmt.Printf("Конфликт секрета логина/пароля %s:\n", secret.SecretName)
			fmt.Println("Локальная версия:")
			fmt.Printf("Username: %s\nPassword: %s\n", localSecret.Username, localSecret.Password)
			fmt.Println("Серверная версия:")
			fmt.Printf("Username: %s\nPassword: %s\n", serverSecret.Username, serverSecret.Password)
			fmt.Print("Выберите версию для сохранения (client/server/skip): ")

			var choice string
			_, err = fmt.Scanln(&choice)
			if err != nil {
				fmt.Printf("Ошибка при вводе выбора: %v\n", err)
				continue
			}

			switch choice {
			case "client":
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretUsernamePasswordGRPC(ctx, cl, token, saveReq); err != nil {
					fmt.Printf("Ошибка при сохранении локального секрета логина/пароля %s на сервере: %v\n", secret.SecretName, err)
				}
			case "server":
				continue
			case "skip":
				fmt.Println("Пропуск секрета.")
			default:
				fmt.Println("Неверный выбор. Пропуск секрета.")
			}
		}
	}
	return nil
}

func syncBankCardHTTP(
	ctx context.Context,
	c *resty.Client,
	db *sqlx.DB,
	token string,
	resolve string,
	localSecrets []models.SecretBankCardGetRequest,
) error {
	makeSaveReq := func(secret *models.SecretBankCardGetResponse) models.SecretBankCardSaveRequest {
		return models.SecretBankCardSaveRequest{
			SecretName: secret.SecretName,
			Number:     secret.Number,
			Owner:      secret.Owner,
			Exp:        secret.Exp,
			CVV:        secret.CVV,
			Meta:       secret.Meta,
		}
	}

	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			fullSecret, err := client.GetSecretBankCardByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный полный банковский секрет %s: %v\n", secret.SecretName, err)
				continue
			}
			saveReq := makeSaveReq(fullSecret)
			if err := client.SaveSecretBankCardHTTP(ctx, c, token, saveReq); err != nil {
				fmt.Printf("Ошибка при отправке банковского секрета %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		// Логика загрузки с сервера при необходимости
		return nil
	case "interactive":
		for _, secret := range localSecrets {
			localSecret, err := client.GetSecretBankCardByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный банковский секрет %s: %v\n", secret.SecretName, err)
				continue
			}

			serverSecret, err := client.GetSecretBankCardHTTP(ctx, c, token, secret.SecretName)
			if err != nil {
				fmt.Printf("Серверный банковский секрет %s не найден.\n", secret.SecretName)
				fmt.Print("Отправить локальную версию на сервер? (yes/no): ")

				var choice string
				_, err := fmt.Scanln(&choice)
				if err != nil {
					fmt.Printf("Ошибка при вводе выбора: %v\n", err)
					continue
				}
				if choice == "yes" {
					saveReq := makeSaveReq(localSecret)
					if err := client.SaveSecretBankCardHTTP(ctx, c, token, saveReq); err != nil {
						fmt.Printf("Ошибка при отправке банковского секрета %s: %v\n", localSecret.SecretName, err)
					}
				}
				continue
			}

			fmt.Printf("Конфликт банковского секрета %s:\n", secret.SecretName)
			fmt.Println("Локальная версия:")
			fmt.Printf("Number: %s\nOwner: %s\nExp: %s\nCVV: %s\n", localSecret.Number, localSecret.Owner, localSecret.Exp, localSecret.CVV)
			fmt.Println("Серверная версия:")
			fmt.Printf("Number: %s\nOwner: %s\nExp: %s\nCVV: %s\n", serverSecret.Number, serverSecret.Owner, serverSecret.Exp, serverSecret.CVV)
			fmt.Print("Выберите версию для сохранения (client/server/skip): ")

			var choice string
			_, err = fmt.Scanln(&choice)
			if err != nil {
				fmt.Printf("Ошибка при вводе выбора: %v\n", err)
				continue
			}

			switch choice {
			case "client":
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretBankCardHTTP(ctx, c, token, saveReq); err != nil {
					fmt.Printf("Ошибка при сохранении локального банковского секрета %s на сервере: %v\n", secret.SecretName, err)
				}
			case "server":
				// Серверная версия актуальна — ничего не делаем
			case "skip":
				fmt.Println("Пропуск секрета.")
			default:
				fmt.Println("Неверный выбор. Пропуск секрета.")
			}
		}
	}
	return nil
}

func syncBankCardGRPC(
	ctx context.Context,
	grpcClient grpc.ClientConnInterface,
	db *sqlx.DB,
	token string,
	resolve string,
	localSecrets []models.SecretBankCardGetRequest,
) error {
	cl := pb.NewSecretBankCardServiceClient(grpcClient)

	makeSaveReq := func(secret *models.SecretBankCardGetResponse) models.SecretBankCardSaveRequest {
		return models.SecretBankCardSaveRequest{
			SecretName: secret.SecretName,
			Number:     secret.Number,
			Owner:      secret.Owner,
			Exp:        secret.Exp,
			CVV:        secret.CVV,
			Meta:       secret.Meta,
		}
	}

	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			fullSecret, err := client.GetSecretBankCardByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный полный банковский секрет %s: %v\n", secret.SecretName, err)
				continue
			}
			saveReq := makeSaveReq(fullSecret)
			if err := client.SaveSecretBankCardGRPC(ctx, cl, token, saveReq); err != nil {
				fmt.Printf("Ошибка при отправке банковского секрета %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		return nil
	case "interactive":
		for _, secret := range localSecrets {
			localSecret, err := client.GetSecretBankCardByNameRequest(ctx, db, secret.SecretName)
			if err != nil {
				fmt.Printf("Не удалось получить локальный банковский секрет %s: %v\n", secret.SecretName, err)
				continue
			}

			serverSecret, err := client.GetSecretBankCardGRPC(ctx, cl, token, secret.SecretName)
			if err != nil {
				fmt.Printf("Серверный банковский секрет %s не найден.\n", secret.SecretName)
				fmt.Print("Отправить локальную версию на сервер? (yes/no): ")

				var choice string
				_, err := fmt.Scanln(&choice)
				if err != nil {
					fmt.Printf("Ошибка при вводе выбора: %v\n", err)
					continue
				}
				if choice == "yes" {
					saveReq := makeSaveReq(localSecret)
					if err := client.SaveSecretBankCardGRPC(ctx, cl, token, saveReq); err != nil {
						fmt.Printf("Ошибка при отправке банковского секрета %s: %v\n", localSecret.SecretName, err)
					}
				}
				continue
			}

			fmt.Printf("Конфликт банковского секрета %s:\n", secret.SecretName)
			fmt.Println("Локальная версия:")
			fmt.Printf("Number: %s\nOwner: %s\nExp: %s\nCVV: %s\n", localSecret.Number, localSecret.Owner, localSecret.Exp, localSecret.CVV)
			fmt.Println("Серверная версия:")
			fmt.Printf("Number: %s\nOwner: %s\nExp: %s\nCVV: %s\n", serverSecret.Number, serverSecret.Owner, serverSecret.Exp, serverSecret.CVV)
			fmt.Print("Выберите версию для сохранения (client/server/skip): ")

			var choice string
			_, err = fmt.Scanln(&choice)
			if err != nil {
				fmt.Printf("Ошибка при вводе выбора: %v\n", err)
				continue
			}

			switch choice {
			case "client":
				saveReq := makeSaveReq(localSecret)
				if err := client.SaveSecretBankCardGRPC(ctx, cl, token, saveReq); err != nil {
					fmt.Printf("Ошибка при сохранении локального банковского секрета %s на сервере: %v\n", secret.SecretName, err)
				}
			case "server":
				// Серверная версия актуальна
			case "skip":
				fmt.Println("Пропуск секрета.")
			default:
				fmt.Println("Неверный выбор. Пропуск секрета.")
			}
		}
	}
	return nil
}
