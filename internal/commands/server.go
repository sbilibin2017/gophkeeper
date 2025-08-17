package commands

import (
	"context"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/address"
	"github.com/spf13/cobra"
)

func NewServerCommand(
	runHTTP func(
		ctx context.Context,
		serverURL string,
		databaseDriver string,
		databaseDSN string,
		databaseMaxOpenConns int,
		databaseMaxIdleConns int,
		databaseConnMaxLifetime time.Duration,
		migrationsDir string,
		jwtSecret string,
		jwtExp time.Duration,
	) error,
) *cobra.Command {

	var (
		serverURL               string
		databaseDriver          string
		databaseDSN             string
		databaseMaxOpenConns    int
		databaseMaxIdleConns    int
		databaseConnMaxLifetime time.Duration
		migrationsDir           string
		jwtSecret               string
		jwtExp                  time.Duration
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Запускает HTTP сервер GophKeeper",
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := address.New(serverURL)

			switch addr.Scheme {
			case address.SchemeHTTP, address.SchemeHTTPS:
				return runHTTP(
					cmd.Context(),
					addr.String(),
					databaseDriver,
					databaseDSN,
					databaseMaxOpenConns,
					databaseMaxIdleConns,
					databaseConnMaxLifetime,
					migrationsDir,
					jwtSecret,
					jwtExp,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "localhost:8080", "URL для запуска сервера")
	cmd.Flags().StringVar(&databaseDriver, "database-driver", "postgres", "Драйвер базы данных")
	cmd.Flags().StringVar(&databaseDSN, "database-dsn", "", "DSN для подключения к базе данных")
	cmd.Flags().IntVar(&databaseMaxOpenConns, "database-max-open-conns", 10, "Максимальное количество открытых соединений к базе")
	cmd.Flags().IntVar(&databaseMaxIdleConns, "database-max-idle-conns", 5, "Максимальное количество простаивающих соединений")
	cmd.Flags().DurationVar(&databaseConnMaxLifetime, "database-conn-max-lifetime", time.Hour, "Максимальное время жизни соединения к базе")
	cmd.Flags().StringVar(&migrationsDir, "migrations-dir", "migrations", "Папка с миграциями базы данных")
	cmd.Flags().StringVar(&jwtSecret, "jwt-secret", "secret", "Секретный ключ для JWT")
	cmd.Flags().DurationVar(&jwtExp, "jwt-exp", 24*time.Hour, "Время жизни JWT токена")

	return cmd
}
