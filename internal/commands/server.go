package commands

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/internal/configs/address"
)

func NewServerCommand(
	httpRunner func(
		ctx context.Context,
		serverURL string,
		databaseDSN string,
		databaseMigrationsDir string,
		jwtSecretKey string,
		jwtExp time.Duration,
	) error,
) *cobra.Command {
	var (
		serverAddr            string
		databaseDSN           string
		databaseMigrationsDir string
		jwtSecretKey          string
		jwtExp                time.Duration
	)

	cmd := &cobra.Command{
		Use:     "сервер",
		Short:   "Запуск HTTP сервера GophKeeper",
		Example: "gophkeeper-server --address :8080 --database-dsn gophkeeper.db --jwt-secret secret",
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := address.New(serverAddr)
			if err != nil {
				return err
			}

			switch addr.Scheme {
			default:
				return httpRunner(
					cmd.Context(),
					addr.Address,
					databaseDSN,
					databaseMigrationsDir,
					jwtSecretKey,
					jwtExp,
				)
			}
		},
	}

	cmd.Flags().StringVarP(&serverAddr, "address", "a", ":8080", "адрес сервера (host:port или http://host:port)")
	cmd.Flags().StringVarP(&databaseDSN, "database-dsn", "d", "gophkeeper.db", "DSN базы данных")
	cmd.Flags().StringVarP(&databaseMigrationsDir, "migrations-dir", "m", "migrations", "директория с миграциями базы данных")
	cmd.Flags().StringVarP(&jwtSecretKey, "jwt-secret", "s", "secret", "секретный ключ для JWT")
	cmd.Flags().DurationVarP(&jwtExp, "jwt-exp", "e", time.Hour, "время жизни JWT")

	return cmd
}
