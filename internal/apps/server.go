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
	// создаём подключение к БД
	conn, err := db.New(databaseDriver, databaseDSN,
		db.WithMaxOpenConns(databaseMaxOpenConns),
		db.WithMaxIdleConns(databaseMaxIdleConns),
		db.WithConnMaxLifetime(databaseConnMaxLifetime),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	// выполняем миграции
	if err := db.RunMigrations(conn, databaseDriver, migrationsDir); err != nil {
		return err
	}

	// инициализируем репозитории
	userReadRepo := repositories.NewUserReadRepository(conn)
	userWriteRepo := repositories.NewUserWriteRepository(conn)
	deviceWriteRepo := repositories.NewDeviceWriteRepository(conn)

	// инициализируем JWT
	jwt := jwt.New(jwtSecret, jwtExp)

	// инициализируем rsa
	rsa := rsa.New()

	// инициализируем hasher
	pwHasher := passwordhasher.New()

	// хендлер регистрации
	registerHTTPHandler := handlers.NewRegisterHTTPHandler(
		userReadRepo,
		userWriteRepo,
		deviceWriteRepo,
		jwt,
		rsa,
		pwHasher,
		validators.ValidateUsername,
		validators.ValidatePassword,
	)

	// роутер
	router := chi.NewRouter()
	router.Post("/register", registerHTTPHandler)

	// создаём HTTP сервер
	srv := &http.Server{
		Addr:    serverURL,
		Handler: router,
	}

	// создаём контекст, который реагирует на сигналы завершения
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	// запуск сервера в отдельной горутине
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// ждём либо сигнала завершения, либо ошибки сервера
	select {
	case <-ctx.Done():
		// graceful shutdown с таймаутом 5 секунд
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
