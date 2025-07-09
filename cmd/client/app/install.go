package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"

	"github.com/spf13/cobra"
)

// newInstallCommand создаёт новую команду "install" для CLI,
// которая позволяет установить клиент для текущей платформы.
// Команда принимает флаг "server-url" для указания базового URL сервера,
// с которого будет скачиваться клиент.
func newInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Установить клиент для текущей платформы",
		Long: `Команда для установки клиента приложения на вашу текущую операционную систему и архитектуру.
Клиент будет скачан с указанного сервера, заданного через флаг --server-url.
Флаг --server-url обязателен и должен содержать адрес HTTP(S) или gRPC сервера.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			serverURL, err := parseInstallFlags(cmd)
			if err != nil {
				return err
			}

			return runInstallApp(serverURL)
		},
	}

	cmd.Flags().String("server-url", "", "Базовый URL сервера для скачивания клиента")

	_ = cmd.MarkFlagRequired("server-url")

	return cmd
}

// parseInstallFlags парсит и проверяет обязательные флаги команды install.
// Возвращает ошибку, если флаг server-url не указан или произошла ошибка при его чтении.
// parseInstallFlags парсит и проверяет обязательные флаги команды install.
// Возвращает ошибку, если флаг server-url не указан или произошла ошибка при его чтении.
func parseInstallFlags(cmd *cobra.Command) (string, error) {
	serverURL, err := cmd.Flags().GetString("server-url")
	if err != nil {
		return "", fmt.Errorf("не удалось прочитать флаг server-url: %w", err)
	}

	if serverURL == "" {
		return "", fmt.Errorf("флаг server-url обязателен для установки клиента")
	}

	return serverURL, nil
}

// runInstallApp выполняет основную логику установки клиента.
// Создаёт конфигурацию клиента с HTTP и gRPC клиентами,
// пытается установить клиента сначала через HTTP, если доступен, иначе через gRPC.
// Возвращает ошибку в случае неудачи.
func runInstallApp(serverURL string) error {
	var (
		cfg *configs.ClientConfig
		err error
	)

	switch {
	case strings.HasPrefix(serverURL, "http://"), strings.HasPrefix(serverURL, "https://"):
		cfg, err = configs.NewClientConfig(configs.WithHTTPClient(serverURL))
	case strings.HasPrefix(serverURL, "grpc://"):
		cfg, err = configs.NewClientConfig(configs.WithGRPCClient(serverURL))
	default:
		return fmt.Errorf("протокол в URL сервера не поддерживается: %s", serverURL)
	}

	if err != nil {
		return fmt.Errorf("ошибка при создании конфигурации клиента: %w", err)
	}
	defer func() {
		if cfg.GRPCClient != nil {
			cfg.GRPCClient.Close()
		}
	}()

	ctx := context.Background()

	if cfg.HTTPClient != nil {
		if err := services.ClientInstallHTTP(ctx, cfg.HTTPClient); err != nil {
			return fmt.Errorf("ошибка установки клиента через HTTP: %w", err)
		}
		return nil
	}

	if cfg.GRPCClient != nil {
		client := pb.NewClientInstallServiceClient(cfg.GRPCClient)
		if err := services.ClientInstallGRPC(ctx, client); err != nil {
			return fmt.Errorf("ошибка установки клиента через gRPC: %w", err)
		}
		return nil
	}

	return fmt.Errorf("нет доступных клиентов для установки")
}
