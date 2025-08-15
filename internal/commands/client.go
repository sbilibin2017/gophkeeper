package commands

import (
	"context"

	"github.com/spf13/cobra"
)

// NewRootCommand создает корневую CLI-команду для клиента GophKeeper.
func NewRootCommand() *cobra.Command {
	return &cobra.Command{Use: "gophkeeper-client"}
}

// NewRegisterCommand создает CLI-команду для регистрации пользователя и устройства.
func NewRegisterCommand(
	httpRunner func(
		ctx context.Context,
		serverURL string,
		databaseMigrationsDir string,
		username string,
		password string,
		deviceID string,
	) ([]byte, string, error),
	deviceIDGetter func() (string, error),
) *cobra.Command {
	var (
		serverURL string
		username  string
		password  string
	)

	cmd := &cobra.Command{
		Use:     "register",
		Short:   "Регистрация нового пользователя и устройства GophKeeper",
		Example: "gophkeeper-client register --username user1 --password secret123 --server-url http://localhost:8080",
		RunE: func(cmd *cobra.Command, args []string) error {
			deviceID, err := deviceIDGetter()
			if err != nil {
				return err
			}
			privKey, token, err := httpRunner(
				cmd.Context(),
				serverURL,
				"migrations",
				username,
				password,
				deviceID,
			)
			if err != nil {
				return err
			}

			cmd.Println("Регистрация успешна")
			cmd.Printf("Приватный ключ: %s\n", string(privKey))
			cmd.Printf("Токен: %s\n", token)

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "URL сервера GophKeeper")
	cmd.Flags().StringVar(&username, "username", "", "имя пользователя для регистрации")
	cmd.Flags().StringVar(&password, "password", "", "пароль для регистрации")

	return cmd
}
