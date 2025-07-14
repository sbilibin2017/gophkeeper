package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
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
			switch resolve {
			case "client", "server", "interactive":
			default:
				return fmt.Errorf("неподдерживаемая стратегия resolve: %s (возможные: client, server, interactive)", resolve)
			}

			if serverURL == "" {
				return fmt.Errorf("не указан адрес сервера (--server-url)")
			}

			ctx := context.Background()

			var config *configs.ClientConfig
			var err error

			if strings.HasPrefix(serverURL, "http://") || strings.HasPrefix(serverURL, "https://") {
				config, err = configs.NewClientConfig(
					configs.WithHTTPClient(serverURL),
					configs.WithDB("gophkeeper_client.db"),
				)
			} else if strings.HasPrefix(serverURL, "grpc://") {
				config, err = configs.NewClientConfig(
					configs.WithGRPCClient(serverURL),
					configs.WithDB("gophkeeper_client.db"),
				)
			} else {
				return fmt.Errorf("неподдерживаемый префикс сервера: %s", serverURL)
			}

			if err != nil {
				return fmt.Errorf("не удалось создать конфиг клиента: %w", err)
			}

			// Репозитории
			textRepo := repositories.NewSecretTextClientListRepository(config.DB)
			binaryRepo := repositories.NewSecretBinaryClientListRepository(config.DB)
			bankCardRepo := repositories.NewSecretBankCardClientListRepository(config.DB)
			loginPasswordRepo := repositories.NewSecretUsernamePasswordClientListRepository(config.DB)

			// Получаем локальные секреты
			textSecrets, err := textRepo.List(ctx)
			if err != nil {
				return fmt.Errorf("ошибка получения текстовых секретов: %w", err)
			}

			binarySecrets, err := binaryRepo.List(ctx)
			if err != nil {
				return fmt.Errorf("ошибка получения бинарных секретов: %w", err)
			}

			bankCardSecrets, err := bankCardRepo.List(ctx)
			if err != nil {
				return fmt.Errorf("ошибка получения банковских карт: %w", err)
			}

			loginPasswordSecrets, err := loginPasswordRepo.List(ctx)
			if err != nil {
				return fmt.Errorf("ошибка получения логинов и паролей: %w", err)
			}

			if strings.HasPrefix(serverURL, "http://") || strings.HasPrefix(serverURL, "https://") {
				client := config.HTTPClient

				textSaveFacade := facades.NewSecretTextSaveHTTPFacade(client)
				textGetFacade := facades.NewSecretTextGetHTTPFacade(client)

				binarySaveFacade := facades.NewSecretBinarySaveHTTPFacade(client)
				binaryGetFacade := facades.NewSecretBinaryGetHTTPFacade(client)

				bankCardSaveFacade := facades.NewSecretBankCardSaveHTTPFacade(client)
				bankCardGetFacade := facades.NewSecretBankCardGetHTTPFacade(client)

				loginPasswordSaveFacade := facades.NewSecretUsernamePasswordSaveHTTPFacade(client)
				loginPasswordGetFacade := facades.NewSecretUsernamePasswordGetHTTPFacade(client)

				if err := syncBankCardSecrets(ctx, bankCardSecrets, bankCardSaveFacade, bankCardGetFacade, token, resolve); err != nil {
					return err
				}

				if err := syncTextSecrets(ctx, textSecrets, textSaveFacade, textGetFacade, token, resolve); err != nil {
					return err
				}

				if err := syncBinarySecrets(ctx, binarySecrets, binarySaveFacade, binaryGetFacade, token, resolve); err != nil {
					return err
				}

				if err := syncLoginPasswordSecrets(ctx, loginPasswordSecrets, loginPasswordSaveFacade, loginPasswordGetFacade, token, resolve); err != nil {
					return err
				}

			} else {
				client := config.GRPCClient

				textSaveFacade := facades.NewSecretTextSaveGRPCFacade(grpc.NewSecretTextServiceClient(client))
				textGetFacade := facades.NewSecretTextGetGRPCFacade(grpc.NewSecretTextServiceClient(client))

				binarySaveFacade := facades.NewSecretBinarySaveGRPCFacade(grpc.NewSecretBinaryServiceClient(client))
				binaryGetFacade := facades.NewSecretBinaryGetGRPCFacade(grpc.NewSecretBinaryServiceClient(client))

				bankCardSaveFacade := facades.NewSecretBankCardSaveGRPCFacade(grpc.NewSecretBankCardServiceClient(client))
				bankCardGetFacade := facades.NewSecretBankCardGetGRPCFacade(grpc.NewSecretBankCardServiceClient(client))

				loginPasswordSaveFacade := facades.NewSecretUsernamePasswordSaveGRPCFacade(grpc.NewSecretUsernamePasswordServiceClient(client))
				loginPasswordGetFacade := facades.NewSecretUsernamePasswordGetGRPCFacade(grpc.NewSecretUsernamePasswordServiceClient(client))

				if err := syncBankCardSecrets(ctx, bankCardSecrets, bankCardSaveFacade, bankCardGetFacade, token, resolve); err != nil {
					return err
				}

				if err := syncTextSecrets(ctx, textSecrets, textSaveFacade, textGetFacade, token, resolve); err != nil {
					return err
				}

				if err := syncBinarySecrets(ctx, binarySecrets, binarySaveFacade, binaryGetFacade, token, resolve); err != nil {
					return err
				}

				if err := syncLoginPasswordSecrets(ctx, loginPasswordSecrets, loginPasswordSaveFacade, loginPasswordGetFacade, token, resolve); err != nil {
					return err
				}

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

func syncTextSecrets(
	ctx context.Context,
	localSecrets []models.SecretTextClient,
	saveClient interface {
		Save(ctx context.Context, token string, secret models.SecretTextClient) error
	},
	getter interface {
		Get(ctx context.Context, token string, secretName string) (*models.SecretTextClient, error)
	},
	token string,
	resolve string,
) error {
	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			if err := saveClient.Save(ctx, token, secret); err != nil {
				fmt.Printf("Ошибка при отправке текстового секрета %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		return nil
	case "interactive":
		for _, localSecret := range localSecrets {
			serverSecret, err := getter.Get(ctx, token, localSecret.SecretName)
			if err != nil {
				if err := saveClient.Save(ctx, token, localSecret); err != nil {
					fmt.Printf("Ошибка при отправке текстового секрета %s: %v\n", localSecret.SecretName, err)
				}
				continue
			}
			if serverSecret.UpdatedAt.Before(localSecret.UpdatedAt) {
				saveLocal, err := chooseVersion(localSecret.SecretName, localSecret, *serverSecret)
				if err != nil {
					fmt.Println("Пропускаем секрет из-за ошибки выбора.")
					continue
				}
				if saveLocal {
					if err := saveClient.Save(ctx, token, localSecret); err != nil {
						fmt.Printf("Ошибка при отправке текстового секрета %s: %v\n", localSecret.SecretName, err)
					}
				}
			}
		}
	}
	return nil
}

func syncBinarySecrets(
	ctx context.Context,
	localSecrets []models.SecretBinaryClient,
	saveClient interface {
		Save(ctx context.Context, token string, secret models.SecretBinaryClient) error
	},
	getter interface {
		Get(ctx context.Context, token string, secretName string) (*models.SecretBinaryClient, error)
	},
	token string,
	resolve string,
) error {
	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			if err := saveClient.Save(ctx, token, secret); err != nil {
				fmt.Printf("Ошибка при отправке бинарного секрета %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		return nil
	case "interactive":
		for _, localSecret := range localSecrets {
			serverSecret, err := getter.Get(ctx, token, localSecret.SecretName)
			if err != nil {
				if err := saveClient.Save(ctx, token, localSecret); err != nil {
					fmt.Printf("Ошибка при отправке бинарного секрета %s: %v\n", localSecret.SecretName, err)
				}
				continue
			}
			if serverSecret.UpdatedAt.Before(localSecret.UpdatedAt) {
				saveLocal, err := chooseVersion(localSecret.SecretName, localSecret, *serverSecret)
				if err != nil {
					fmt.Println("Пропускаем секрет из-за ошибки выбора.")
					continue
				}
				if saveLocal {
					if err := saveClient.Save(ctx, token, localSecret); err != nil {
						fmt.Printf("Ошибка при отправке бинарного секрета %s: %v\n", localSecret.SecretName, err)
					}
				}
			}
		}
	}
	return nil
}

func syncLoginPasswordSecrets(
	ctx context.Context,
	localSecrets []models.SecretUsernamePasswordClient,
	saveClient interface {
		Save(ctx context.Context, token string, secret models.SecretUsernamePasswordClient) error
	},
	getter interface {
		Get(ctx context.Context, token string, secretName string) (*models.SecretUsernamePasswordClient, error)
	},
	token string,
	resolve string,
) error {
	switch resolve {
	case "client":
		for _, secret := range localSecrets {
			if err := saveClient.Save(ctx, token, secret); err != nil {
				fmt.Printf("Ошибка при отправке логина/пароля %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		return nil
	case "interactive":
		for _, localSecret := range localSecrets {
			serverSecret, err := getter.Get(ctx, token, localSecret.SecretName)
			if err != nil {
				if err := saveClient.Save(ctx, token, localSecret); err != nil {
					fmt.Printf("Ошибка при отправке логина/пароля %s: %v\n", localSecret.SecretName, err)
				}
				continue
			}
			if serverSecret.UpdatedAt.Before(localSecret.UpdatedAt) {
				saveLocal, err := chooseVersion(localSecret.SecretName, localSecret, *serverSecret)
				if err != nil {
					fmt.Println("Пропускаем секрет из-за ошибки выбора.")
					continue
				}
				if saveLocal {
					if err := saveClient.Save(ctx, token, localSecret); err != nil {
						fmt.Printf("Ошибка при отправке логина/пароля %s: %v\n", localSecret.SecretName, err)
					}
				}
			}
		}
	}
	return nil
}

func syncBankCardSecrets(
	ctx context.Context,
	localSecrets []models.SecretBankCardClient,
	saveClient interface {
		Save(ctx context.Context, token string, secret models.SecretBankCardClient) error
	},
	getter interface {
		Get(ctx context.Context, token string, secretName string) (*models.SecretBankCardClient, error)
	},
	token string,
	resolve string,
) error {
	switch resolve {
	case "client":
		// Отправляем локальные данные на сервер
		for _, secret := range localSecrets {
			if err := saveClient.Save(ctx, token, secret); err != nil {
				fmt.Printf("Ошибка при отправке банковской карты %s: %v\n", secret.SecretName, err)
			}
		}
	case "server":
		return nil
	case "interactive":
		for _, localSecret := range localSecrets {
			serverSecret, err := getter.Get(ctx, token, localSecret.SecretName)
			if err != nil {
				// Серверная версия отсутствует — отправляем локальную
				if err := saveClient.Save(ctx, token, localSecret); err != nil {
					fmt.Printf("Ошибка при отправке банковской карты %s: %v\n", localSecret.SecretName, err)
				}
				continue
			}

			if serverSecret.UpdatedAt.Before(localSecret.UpdatedAt) {
				saveLocal, err := chooseVersion(localSecret.SecretName, localSecret, *serverSecret)
				if err != nil {
					fmt.Println("Пропускаем секрет из-за ошибки выбора.")
					continue
				}
				if saveLocal {
					if err := saveClient.Save(ctx, token, localSecret); err != nil {
						fmt.Printf("Ошибка при отправке банковской карты %s: %v\n", localSecret.SecretName, err)
					}
				}
			}
		}
	}
	return nil
}

func chooseVersion(name string, localSecret, serverSecret interface{}) (bool, error) {
	fmt.Printf("Конфликт по секрету '%s':\n", name)
	fmt.Println("1) Клиентская версия:", localSecret)
	fmt.Println("2) Серверная версия:", serverSecret)
	fmt.Print("Выберите версию для сохранения (1 - клиентская, 2 - серверная): ")

	var choice int
	_, err := fmt.Scan(&choice)
	if err != nil || (choice != 1 && choice != 2) {
		return false, fmt.Errorf("некорректный выбор")
	}

	return choice == 1, nil
}
