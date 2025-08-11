package server

import (
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/address"
	"github.com/spf13/cobra"
)

// NewCommand creates a new Cobra command to start the registration server.
func NewCommand() *cobra.Command {
	var (
		serverURL    string
		databaseDSN  string
		jwtSecretKey string
		jwtExp       time.Duration

		apiVersion       = "/api/v1"
		databaseDriver   = "sqlite"
		maxOpenConns     = 1
		maxIdleConns     = 1
		connMaxLifetime  = time.Minute
		pathToMigrations = "./migrations"
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Start the registration server (HTTP or gRPC)",
		Long: `Starts the backend registration server.
Use --server-url scheme to decide HTTP or gRPC server.

Example:
  gophkeeper register --server-url http://localhost:8080
  gophkeeper register --server-url grpc://localhost:50051
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := address.New(serverURL)

			switch addr.Scheme {
			case address.SchemeHTTP, address.SchemeHTTPS:
				return RunHTTP(
					cmd.Context(),
					apiVersion,
					databaseDriver,
					databaseDSN,
					maxOpenConns,
					maxIdleConns,
					connMaxLifetime,
					jwtSecretKey,
					jwtExp,
					pathToMigrations,
					addr.Address,
				)
			case address.SchemeGRPC:
				return RunGRPC(
					cmd.Context(),
					databaseDriver,
					databaseDSN,
					maxOpenConns,
					maxIdleConns,
					connMaxLifetime,
					jwtSecretKey,
					jwtExp,
					pathToMigrations,
					addr.Address,
				)
			default:
				return address.ErrUnsupportedScheme
			}
		},
	}

	cmd.Flags().StringVar(&serverURL, "server-url", "http://localhost:8080", "Server URL with scheme (http:// or grpc://)")
	cmd.Flags().StringVar(&databaseDSN, "database-dsn", "", "Database DSN")
	cmd.Flags().StringVar(&jwtSecretKey, "jwt-secret", "", "JWT secret key")
	cmd.Flags().DurationVar(&jwtExp, "jwt-exp", time.Hour, "JWT expiration duration")

	return cmd
}
