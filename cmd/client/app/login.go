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
	loginUsername string // Имя пользователя для аутентификации
	loginPassword string // Пароль пользователя
)

func newLoginCommand() *cobra.Command {
	var (
		serverURL   string
		interactive bool
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Аутентификация пользователя",
		Long: `Команда для аутентификации пользователя в системе Gophkeeper.

Поддерживается передача имени пользователя, пароля и URL сервера как через флаги,
так и через интерактивный ввод. Если URL сервера не указан, используется
значение из переменной окружения GOPHKEEPER_SERVER_URL.`,
		Example: `  gophkeeper login --username alice --password secret123 --server-url https://example.com
  gophkeeper login --interactive`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := parseLoginFlags(&serverURL, &interactive); err != nil {
				return err
			}

			opts, err := options.NewOptions(
				options.WithServerURL(serverURL),
			)
			if err != nil {
				return fmt.Errorf("не удалось создать конфигурацию клиента: %w", err)
			}

			if opts.ClientConfig.HTTPClient != nil {
				token, err := services.LoginHTTP(
					context.Background(),
					opts.ClientConfig.HTTPClient,
					loginUsername,
					loginPassword,
				)
				if err != nil {
					return err
				}
				fmt.Print(token)
				return nil
			}

			if opts.ClientConfig.GRPCClient != nil {
				defer opts.ClientConfig.GRPCClient.Close()

				client := pb.NewLoginServiceClient(opts.ClientConfig.GRPCClient)
				token, err := services.LoginGRPC(
					context.Background(),
					client,
					loginUsername,
					loginPassword,
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

	cmd.Flags().StringVar(&loginUsername, "username", "", "Имя пользователя для аутентификации")
	cmd.Flags().StringVar(&loginPassword, "password", "", "Пароль пользователя")

	cmd = options.RegisterServerURLFlag(cmd, &serverURL)
	cmd = options.RegisterInteractiveFlag(cmd, &interactive)

	return cmd
}

func parseLoginFlags(serverURL *string, interactive *bool) error {
	if *interactive {
		reader := bufio.NewReader(os.Stdin)
		if err := parseLoginFlagsInteractive(reader, serverURL); err != nil {
			return err
		}
	}

	if loginUsername == "" || loginPassword == "" {
		return fmt.Errorf("имя пользователя и пароль не могут быть пустыми")
	}

	if *serverURL == "" {
		*serverURL = os.Getenv("GOPHKEEPER_SERVER_URL")
	}

	if *serverURL == "" {
		return fmt.Errorf("URL сервера не может быть пустым")
	}

	return nil
}

func parseLoginFlagsInteractive(r *bufio.Reader, serverURL *string) error {
	fmt.Print("Введите имя пользователя: ")
	userInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	loginUsername = strings.TrimSpace(userInput)

	fmt.Print("Введите пароль: ")
	passInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	loginPassword = strings.TrimSpace(passInput)

	fmt.Print("Введите URL сервера (оставьте пустым для использования GOPHKEEPER_SERVER_URL из окружения): ")
	urlInput, err := r.ReadString('\n')
	if err != nil {
		return err
	}
	*serverURL = strings.TrimSpace(urlInput)

	return nil
}
