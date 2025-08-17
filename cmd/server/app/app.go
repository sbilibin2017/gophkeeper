package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/gophkeeper/internal/configs/address"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/handlers"
	"github.com/sbilibin2017/gophkeeper/internal/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/validators"
	"github.com/spf13/cobra"
)

// NewCommand создаёт и возвращает новую команду Cobra для запуска HTTP-сервера GophKeeper.
func NewCommand() *cobra.Command {
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

	cmd.Flags().StringVar(&serverURL, "server-url", ":8080", "URL для запуска сервера")
	cmd.Flags().StringVar(&databaseDriver, "database-driver", "sqlite", "Драйвер базы данных")
	cmd.Flags().StringVar(&databaseDSN, "database-dsn", "server.db", "DSN для подключения к базе данных")
	cmd.Flags().IntVar(&databaseMaxOpenConns, "database-max-open-conns", 10, "Максимальное количество открытых соединений к базе")
	cmd.Flags().IntVar(&databaseMaxIdleConns, "database-max-idle-conns", 5, "Максимальное количество простаивающих соединений")
	cmd.Flags().DurationVar(&databaseConnMaxLifetime, "database-conn-max-lifetime", time.Hour, "Максимальное время жизни соединения к базе")
	cmd.Flags().StringVar(&migrationsDir, "migrations-dir", "migrations", "Папка с миграциями базы данных")
	cmd.Flags().StringVar(&jwtSecret, "jwt-secret", "secret", "Секретный ключ для JWT")
	cmd.Flags().DurationVar(&jwtExp, "jwt-exp", 24*time.Hour, "Время жизни JWT токена")

	return cmd
}

// runHTTP запускает HTTP-сервер приложения с поддержкой graceful shutdown.
func runHTTP(
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
) error {
	// Инициализируем подключение к БД
	conn, err := db.New(databaseDriver, databaseDSN,
		db.WithMaxOpenConns(databaseMaxOpenConns),
		db.WithMaxIdleConns(databaseMaxIdleConns),
		db.WithConnMaxLifetime(databaseConnMaxLifetime),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Выполняем миграции
	if err := db.RunMigrations(conn, databaseDriver, migrationsDir); err != nil {
		return err
	}

	// Инициализируем репозитории
	userReadRepo := repositories.NewUserReadRepository(conn)
	userWriteRepo := repositories.NewUserWriteRepository(conn)
	deviceWriteRepo := repositories.NewDeviceWriteRepository(conn)
	deviceReadRepo := repositories.NewDeviceReadRepository(conn)
	secretKeyWriteRepo := repositories.NewSecretKeyWriteRepository(conn)
	secretKeyReadRepo := repositories.NewSecretKeyReadRepository(conn)
	secretWriteRepo := repositories.NewSecretWriteRepository(conn)
	secretReadRepo := repositories.NewSecretReadRepository(conn)

	// Инициализируем JWT
	jwt := jwt.New(jwtSecret, jwtExp)

	// Инициализируем роутер
	router := chi.NewRouter()

	// Аутентификация
	router.Post("/register", handlers.NewRegisterHTTPHandler(
		userReadRepo,
		userWriteRepo,
		deviceWriteRepo,
		jwt,
		validators.ValidateUsername,
		validators.ValidatePassword,
	))
	router.Post("/login", handlers.NewLoginHTTPHandler(
		userReadRepo,
		deviceReadRepo,
		jwt,
	))

	// Операции с устройствами
	router.Get("/get-device", handlers.NewDeviceGetHTTPHandler(jwt, deviceReadRepo))

	// Операции с симметричными плючыми
	router.Post("/save-secret-key", handlers.NewSecretKeySaveHTTPHandler(jwt, secretKeyWriteRepo))
	router.Get("/get-secret-key/{secret-id}", handlers.NewSecretKeyGetHTTPHandler(jwt, secretKeyReadRepo))

	// Операции с секретами
	router.Post("/save-secret", handlers.NewSecretSaveHTTPHandler(jwt, secretWriteRepo))
	router.Get("/get-secret/{secret-id}", handlers.NewSecretGetHTTPHandler(jwt, secretReadRepo))
	router.Get("/list-secrets", handlers.NewSecretListHTTPHandler(jwt, secretReadRepo))

	// Инициализируем HTTP сервер
	srv := &http.Server{
		Addr:    serverURL,
		Handler: router,
	}

	// Инициализируемы контекст, который реагирует на сигналы завершения
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// Запуск сервера в отдельной горутине
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// Ждём либо сигнала завершения, либо ошибки сервера
	select {
	case <-ctx.Done():
		// Graceful shutdown с таймаутом 5 секунд
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}
