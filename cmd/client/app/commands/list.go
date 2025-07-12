package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/configs"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

// NewListUsernamePasswordCommand создаёт команду CLI `list-username-password`,
// которая запрашивает и выводит список сохранённых логинов и паролей из сервера.
// Поддерживает как HTTP, так и gRPC, а также опциональный интерактивный режим.
func NewListUsernamePasswordCommand() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:   "list-username-password",
		Short: "Получить список UsernamePassword из сервера",
		Long:  `Получение полного списка UsernamePassword записей из удалённого сервера через HTTP или gRPC.`,
		Args:  cobra.NoArgs,
		Example: `  # Получить список UsernamePassword
  gophkeeper list-username-password

  # В интерактивном режиме
  gophkeeper list-username-password --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			records, err := listUsernamePassword(ctx)
			if err != nil {
				return fmt.Errorf("ошибка получения данных: %w", err)
			}

			if len(records) == 0 {
				fmt.Println("Нет сохранённых логинов и паролей.")
				return nil
			}

			if interactive {
				printUsernamePasswordsInteractive(records)
			} else {
				out, err := json.MarshalIndent(records, "", "  ")
				if err != nil {
					return fmt.Errorf("ошибка при сериализации: %w", err)
				}
				fmt.Println(string(out))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Интерактивный режим показа секретов")
	return cmd
}

// listUsernamePassword получает список логинов и паролей с сервера
// через HTTP или gRPC, используя переменные окружения:
// GOPHKEEPER_SERVER_URL и GOPHKEEPER_TOKEN.
func listUsernamePassword(ctx context.Context) ([]*models.UsernamePassword, error) {
	serverURL := os.Getenv("GOPHKEEPER_SERVER_URL")
	token := os.Getenv("GOPHKEEPER_TOKEN")

	if serverURL == "" {
		return nil, fmt.Errorf("переменная окружения GOPHKEEPER_SERVER_URL не задана")
	}
	if token == "" {
		return nil, fmt.Errorf("переменная окружения GOPHKEEPER_TOKEN не задана")
	}

	if strings.HasPrefix(serverURL, "grpc://") {
		cfg, err := configs.NewClientConfig(configs.WithGRPCClient(serverURL))
		if err != nil {
			return nil, fmt.Errorf("ошибка создания gRPC клиента: %w", err)
		}
		defer func() {
			if cfg.GRPCClient != nil {
				cfg.GRPCClient.Close()
			}
		}()

		grpcClient := pb.NewListServiceClient(cfg.GRPCClient)
		mdCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		grpcItems, err := services.ListUsernamePasswordGRPC(mdCtx, grpcClient)
		if err != nil {
			return nil, err
		}

		var items []*models.UsernamePassword
		for _, i := range grpcItems {
			items = append(items, &models.UsernamePassword{
				Username: i.Username,
				Password: i.Password,
				Meta:     i.Meta,
			})
		}
		return items, nil
	}

	cfg, err := configs.NewClientConfig(configs.WithHTTPClient(serverURL))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP клиента: %w", err)
	}

	cfg.HTTPClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("Authorization", "Bearer "+token)
		return nil
	})

	return services.ListUsernamePasswordHTTP(ctx, cfg.HTTPClient)
}

// printUsernamePasswordsInteractive выводит логины и пароли по одному
// в интерактивном режиме, ожидая нажатие Enter или 'q' для выхода.
func printUsernamePasswordsInteractive(records []*models.UsernamePassword) {
	reader := bufio.NewReader(os.Stdin)

	for i, r := range records {
		fmt.Printf("Секрет #%d:\n", i+1)
		fmt.Printf("  Логин:    %s\n", r.Username)
		fmt.Printf("  Пароль:   %s\n", r.Password)
		if len(r.Meta) > 0 {
			fmt.Println("  Метаданные:")
			for k, v := range r.Meta {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
		fmt.Println(strings.Repeat("-", 30))
		fmt.Print("Нажмите Enter для следующего или 'q' чтобы выйти: ")

		input, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(input)) == "q" {
			fmt.Println("Выход из интерактивного режима.")
			break
		}
	}
}

// --- List Text Command ---

// NewListTextCommand создаёт CLI команду для получения списка текстов
func NewListTextCommand() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:   "list-text",
		Short: "Получить список текстов из сервера",
		Long:  `Получение полного списка текстовых записей из удалённого сервера через HTTP или gRPC.`,
		Args:  cobra.NoArgs,
		Example: `  # Получить список текстов
  gophkeeper list-text

  # В интерактивном режиме
  gophkeeper list-text --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			records, err := listText(ctx)
			if err != nil {
				return fmt.Errorf("ошибка получения данных: %w", err)
			}

			if len(records) == 0 {
				fmt.Println("Нет сохранённых текстов.")
				return nil
			}

			if interactive {
				printTextsInteractive(records)
			} else {
				out, err := json.MarshalIndent(records, "", "  ")
				if err != nil {
					return fmt.Errorf("ошибка сериализации: %w", err)
				}
				fmt.Println(string(out))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Интерактивный режим показа текстов")
	return cmd
}

func listText(ctx context.Context) ([]*models.Text, error) {
	serverURL := os.Getenv("GOPHKEEPER_SERVER_URL")
	token := os.Getenv("GOPHKEEPER_TOKEN")

	if serverURL == "" {
		return nil, fmt.Errorf("переменная окружения GOPHKEEPER_SERVER_URL не задана")
	}
	if token == "" {
		return nil, fmt.Errorf("переменная окружения GOPHKEEPER_TOKEN не задана")
	}

	if strings.HasPrefix(serverURL, "grpc://") {
		cfg, err := configs.NewClientConfig(configs.WithGRPCClient(serverURL))
		if err != nil {
			return nil, fmt.Errorf("ошибка создания gRPC клиента: %w", err)
		}
		defer func() {
			if cfg.GRPCClient != nil {
				cfg.GRPCClient.Close()
			}
		}()

		grpcClient := pb.NewListServiceClient(cfg.GRPCClient)
		mdCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		grpcItems, err := services.ListTextGRPC(mdCtx, grpcClient)
		if err != nil {
			return nil, err
		}

		var items []*models.Text
		for _, i := range grpcItems {
			items = append(items, &models.Text{
				Content: i.Content,
				Meta:    i.Meta,
			})
		}
		return items, nil
	}

	cfg, err := configs.NewClientConfig(configs.WithHTTPClient(serverURL))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP клиента: %w", err)
	}
	cfg.HTTPClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("Authorization", "Bearer "+token)
		return nil
	})

	return services.ListTextHTTP(ctx, cfg.HTTPClient)
}

func printTextsInteractive(records []*models.Text) {
	reader := bufio.NewReader(os.Stdin)

	for i, r := range records {
		fmt.Printf("Текст #%d:\n", i+1)
		fmt.Println(r.Content)
		if len(r.Meta) > 0 {
			fmt.Println("Метаданные:")
			for k, v := range r.Meta {
				fmt.Printf("  %s: %s\n", k, v)
			}
		}
		fmt.Println(strings.Repeat("-", 30))
		fmt.Print("Нажмите Enter для следующего или 'q' чтобы выйти: ")

		input, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(input)) == "q" {
			fmt.Println("Выход из интерактивного режима.")
			break
		}
	}
}

// --- List Binary Command ---

// NewListBinaryCommand создаёт CLI команду для получения списка бинарных данных
func NewListBinaryCommand() *cobra.Command {
	var interactive bool
	var outputDir string

	cmd := &cobra.Command{
		Use:   "list-binary",
		Short: "Получить список бинарных данных из сервера",
		Long:  `Получение полного списка бинарных данных из удалённого сервера через HTTP или gRPC.`,
		Args:  cobra.NoArgs,
		Example: `  # Получить список бинарных данных с сохранением в текущей директории
  gophkeeper list-binary

  # В интерактивном режиме с выбором, сохранять или нет
  gophkeeper list-binary --interactive

  # В неинтерактивном режиме, указать директорию для сохранения
  gophkeeper list-binary --output-dir ./binaries`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			records, err := listBinary(ctx)
			if err != nil {
				return fmt.Errorf("ошибка получения данных: %w", err)
			}

			if len(records) == 0 {
				fmt.Println("Нет сохранённых бинарных данных.")
				return nil
			}

			if interactive {
				printBinariesInteractive(records)
			} else {
				if outputDir == "" {
					outputDir = "." // текущая директория
				}
				err := saveBinariesToDir(records, outputDir)
				if err != nil {
					return fmt.Errorf("ошибка сохранения бинарных данных: %w", err)
				}
				fmt.Printf("Бинарные данные сохранены в директорию: %s\n", outputDir)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Интерактивный режим показа бинарных данных")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Директория для сохранения бинарных данных (неинтерактивный режим)")
	return cmd
}

func listBinary(ctx context.Context) ([]*models.Binary, error) {
	serverURL := os.Getenv("GOPHKEEPER_SERVER_URL")
	token := os.Getenv("GOPHKEEPER_TOKEN")

	if serverURL == "" {
		return nil, fmt.Errorf("переменная окружения GOPHKEEPER_SERVER_URL не задана")
	}
	if token == "" {
		return nil, fmt.Errorf("переменная окружения GOPHKEEPER_TOKEN не задана")
	}

	if strings.HasPrefix(serverURL, "grpc://") {
		cfg, err := configs.NewClientConfig(configs.WithGRPCClient(serverURL))
		if err != nil {
			return nil, fmt.Errorf("ошибка создания gRPC клиента: %w", err)
		}
		defer func() {
			if cfg.GRPCClient != nil {
				cfg.GRPCClient.Close()
			}
		}()

		grpcClient := pb.NewListServiceClient(cfg.GRPCClient)
		mdCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		grpcItems, err := services.ListBinaryGRPC(mdCtx, grpcClient)
		if err != nil {
			return nil, err
		}

		var items []*models.Binary
		for _, i := range grpcItems {
			items = append(items, &models.Binary{
				Data: i.Data,
				Meta: i.Meta,
			})
		}
		return items, nil
	}

	cfg, err := configs.NewClientConfig(configs.WithHTTPClient(serverURL))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP клиента: %w", err)
	}
	cfg.HTTPClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("Authorization", "Bearer "+token)
		return nil
	})

	return services.ListBinaryHTTP(ctx, cfg.HTTPClient)
}

func printBinariesInteractive(records []*models.Binary) {
	reader := bufio.NewReader(os.Stdin)

	for i, r := range records {
		fmt.Printf("Бинарные данные #%d:\n", i+1)
		fmt.Printf("  Размер: %d байт\n", len(r.Data))
		if len(r.Meta) > 0 {
			fmt.Println("  Метаданные:")
			for k, v := range r.Meta {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
		fmt.Print("Сохранить этот файл? (y/n/q): ")

		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "q" {
			fmt.Println("Выход из интерактивного режима.")
			break
		} else if input == "y" {
			// генерируем имя файла с таймштампом
			filename := fmt.Sprintf("binary_%d_%s.bin", i+1, time.Now().Format("20060102150405"))
			err := os.WriteFile(filename, r.Data, 0644)
			if err != nil {
				fmt.Printf("Ошибка сохранения файла: %v\n", err)
			} else {
				fmt.Printf("Файл сохранён как %s\n", filename)
			}
		} else {
			fmt.Println("Файл пропущен.")
		}
		fmt.Println(strings.Repeat("-", 30))
	}
}

func saveBinariesToDir(records []*models.Binary, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for i, r := range records {
		filename := filepath.Join(dir, fmt.Sprintf("binary_%d_%s.bin", i+1, time.Now().Format("20060102150405")))
		if err := os.WriteFile(filename, r.Data, 0644); err != nil {
			return fmt.Errorf("не удалось сохранить файл %s: %w", filename, err)
		}
	}
	return nil
}

// --- List BankCard Command ---

// NewListBankCardCommand создаёт CLI команду для получения списка банковских карт
func NewListBankCardCommand() *cobra.Command {
	var interactive bool

	cmd := &cobra.Command{
		Use:   "list-bank-card",
		Short: "Получить список банковских карт из сервера",
		Long:  `Получение полного списка банковских карт из удалённого сервера через HTTP или gRPC.`,
		Args:  cobra.NoArgs,
		Example: `  # Получить список карт
  gophkeeper list-bank-card

  # В интерактивном режиме
  gophkeeper list-bank-card --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			records, err := listBankCard(ctx)
			if err != nil {
				return fmt.Errorf("ошибка получения данных: %w", err)
			}

			if len(records) == 0 {
				fmt.Println("Нет сохранённых банковских карт.")
				return nil
			}

			if interactive {
				printBankCardsInteractive(records)
			} else {
				out, err := json.MarshalIndent(records, "", "  ")
				if err != nil {
					return fmt.Errorf("ошибка сериализации: %w", err)
				}
				fmt.Println(string(out))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Интерактивный режим показа банковских карт")
	return cmd
}

func listBankCard(ctx context.Context) ([]*models.BankCard, error) {
	serverURL := os.Getenv("GOPHKEEPER_SERVER_URL")
	token := os.Getenv("GOPHKEEPER_TOKEN")

	if serverURL == "" {
		return nil, fmt.Errorf("переменная окружения GOPHKEEPER_SERVER_URL не задана")
	}
	if token == "" {
		return nil, fmt.Errorf("переменная окружения GOPHKEEPER_TOKEN не задана")
	}

	if strings.HasPrefix(serverURL, "grpc://") {
		cfg, err := configs.NewClientConfig(configs.WithGRPCClient(serverURL))
		if err != nil {
			return nil, fmt.Errorf("ошибка создания gRPC клиента: %w", err)
		}
		defer func() {
			if cfg.GRPCClient != nil {
				cfg.GRPCClient.Close()
			}
		}()

		grpcClient := pb.NewListServiceClient(cfg.GRPCClient)
		mdCtx := metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		grpcItems, err := services.ListBankCardGRPC(mdCtx, grpcClient)
		if err != nil {
			return nil, err
		}

		var items []*models.BankCard
		for _, i := range grpcItems {
			items = append(items, &models.BankCard{
				Number: i.Number,
				Owner:  i.Owner,
				Expiry: i.Expiry,
				CVV:    i.Cvv,
				Meta:   i.Meta,
			})
		}
		return items, nil
	}

	cfg, err := configs.NewClientConfig(configs.WithHTTPClient(serverURL))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания HTTP клиента: %w", err)
	}
	cfg.HTTPClient.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		r.SetHeader("Authorization", "Bearer "+token)
		return nil
	})

	return services.ListBankCardHTTP(ctx, cfg.HTTPClient)
}

func printBankCardsInteractive(records []*models.BankCard) {
	reader := bufio.NewReader(os.Stdin)

	for i, r := range records {
		fmt.Printf("Банковская карта #%d:\n", i+1)
		fmt.Printf("  Номер:  %s\n", r.Number)
		fmt.Printf("  Владелец: %s\n", r.Owner)
		fmt.Printf("  Срок действия: %s\n", r.Expiry)
		fmt.Printf("  CVV: %s\n", r.CVV)
		if len(r.Meta) > 0 {
			fmt.Println("  Метаданные:")
			for k, v := range r.Meta {
				fmt.Printf("    %s: %s\n", k, v)
			}
		}
		fmt.Println(strings.Repeat("-", 30))
		fmt.Print("Нажмите Enter для следующей карты или 'q' чтобы выйти: ")

		input, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(input)) == "q" {
			fmt.Println("Выход из интерактивного режима.")
			break
		}
	}
}
