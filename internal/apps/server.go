package apps

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/crypto/passwordhasher"
	"github.com/sbilibin2017/gophkeeper/internal/crypto/rsa"
	"github.com/sbilibin2017/gophkeeper/internal/handlers"
	"github.com/sbilibin2017/gophkeeper/internal/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/validators"
)

// RunServerHTTP запускает HTTP-сервер приложения с поддержкой graceful shutdown.
func RunServerHTTP(
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

	// Инициализируем rsa
	rsa := rsa.New()

	// Инициализируем hasher
	pwHasher := passwordhasher.New()

	// Инициализируем роутер
	router := chi.NewRouter()

	// Аутентификация
	router.Post("/register", handlers.NewRegisterHTTPHandler(
		userReadRepo,
		userWriteRepo,
		deviceWriteRepo,
		jwt,
		rsa,
		pwHasher,
		validators.ValidateUsername,
		validators.ValidatePassword,
	))
	router.Post("/login", handlers.NewLoginHTTPHandler(
		userReadRepo,
		deviceReadRepo,
		jwt,
		pwHasher,
		deviceWriteRepo,
		rsa,
	))

	// Операции с устройствами
	router.Get("/get-device", handlers.NewDeviceGetHTTPHandler(jwt, deviceReadRepo))

	// Операции с симметричными плючыми
	router.Post("/save-secret-key", handlers.NewSecretKeySaveHTTPHandler(jwt, secretKeyWriteRepo))
	router.Get("/get-secret-key", handlers.NewSecretKeyGetHTTPHandler(jwt, secretKeyReadRepo))

	// Операции с секретами
	router.Post("/save-secret", handlers.NewSecretSaveHTTPHandler(jwt, secretWriteRepo))
	router.Get("/get-secret", handlers.NewSecretGetHTTPHandler(jwt, secretReadRepo))
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
