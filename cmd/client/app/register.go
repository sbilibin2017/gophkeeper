package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sbilibin2017/gophkeeper/cmd/client/app/options"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/spf13/cobra"
)

var (
	registerUsername string // Имя пользователя для регистрации
	registerPassword string // Пароль пользователя
)

// newRegisterCommand создаёт команду регистрации нового пользователя.
func newRegisterCommand() *cobra.Command {
	var (
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Регистрация нового пользователя",
		Long: `Команда для регистрации нового пользователя в системе Gophkeeper.

Поддерживается передача имени пользователя, пароля и URL сервера как через флаги,
так и через интерактивный ввод. Если URL сервера не указан, используется
значение из переменной окружения GOPHKEEPER_SERVER_URL.`,
		Example: `  gophkeeper register --username alice --password secret123 --server-url https://example.com
  gophkeeper register --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseRegisterFlags(&serverURL, &interactive); err != nil {
				return err
			}

			opts, err := options.NewOptions(
				options.WithServerURL(serverURL),
			)
			if err != nil {
				return fmt.Errorf("не удалось создать конфигурацию клиента: %w", err)
			}

			if opts.ClientConfig.HTTPClient != nil {
				token, err := services.RegisterHTTP(
					context.Background(),
					opts.ClientConfig.HTTPClient,
					registerUsername,
					registerPassword,
				)
				if err != nil {
					return err
				}
				fmt.Print(token)
				return nil
			}

			if opts.ClientConfig.GRPCClient != nil {
				defer opts.ClientConfig.GRPCClient.Close()

				client := pb.NewRegisterServiceClient(opts.ClientConfig.GRPCClient)
				token, err := services.RegisterGRPC(
					context.Background(),
					client,
					registerUsername,
					registerPassword,
				)
				if err != nil {
					return err
				}
				fmt.Print(token)
				return nil
			}

			return fmt.Errorf("нет доступного клиента (HTTP или gRPC) для подключения")
		},
	}

	// Флаги для имени пользователя и пароля
	cmd.Flags().StringVar(&registerUsername, "username", "", "Имя пользователя для регистрации")
	cmd.Flags().StringVar(&registerPassword, "password", "", "Пароль пользователя")

	// Флаги для URL сервера и интерактивного режима
	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	return cmd
}

// parseRegisterFlags обрабатывает флаги и интерактивный ввод для регистрации.
func parseRegisterFlags(serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseRegisterFlagsInteractive(reader, serverURL); err != nil {
			return err
		}
	}

	// Проверяем, что имя пользователя и пароль заданы
	if registerUsername == "" || registerPassword == "" {
		return fmt.Errorf("имя пользователя и пароль не могут быть пустыми")
	}

	return nil
}

// parseRegisterFlagsInteractive запрашивает данные у пользователя интерактивно.
func parseRegisterFlagsInteractive(r *bufio.Reader, serverURL *string) error {
	fmt.Print("Введите имя пользователя: ")
	userInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	registerUsername = strings.TrimSpace(userInput)

	fmt.Print("Введите пароль: ")
	passInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	registerPassword = strings.TrimSpace(passInput)

	fmt.Print("Введите URL сервера (оставьте пустым для использования GOPHKEEPER_SERVER_URL из окружения): ")
	urlInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*serverURL = strings.TrimSpace(urlInput)

	return nil
}
