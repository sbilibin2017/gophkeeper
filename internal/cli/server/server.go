package server

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/sbilibin2017/gophkeeper/internal/configs/address"
)

func NewCommand(
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
		Use:   "server",
		Short: "Run GophKeeper HTTP server",
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

	cmd.Flags().StringVarP(&serverAddr, "address", "a", ":8080", "server address (host:port or http://host:port)")
	cmd.Flags().StringVarP(&databaseDSN, "database-dsn", "d", "gophkeeper.db", "database DSN")
	cmd.Flags().StringVarP(&databaseMigrationsDir, "migrations-dir", "m", "migrations", "directory with database migrations")
	cmd.Flags().StringVarP(&jwtSecretKey, "jwt-secret", "s", "secret", "JWT secret key")
	cmd.Flags().DurationVarP(&jwtExp, "jwt-exp", "e", time.Hour, "JWT expiration duration")

	return cmd
}
