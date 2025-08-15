package apps

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/pressly/goose/v3"

	"github.com/sbilibin2017/gophkeeper/internal/configs/crypto"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/configs/tx"
	"github.com/sbilibin2017/gophkeeper/internal/handlers"
	"github.com/sbilibin2017/gophkeeper/internal/middlewares"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/sbilibin2017/gophkeeper/internal/validators"
)

// RunServerHTTP запускает HTTP-сервер для работы с API, подключается к базе данных,
// выполняет миграции, инициализирует JWT, сервисы, обработчики и маршрутизатор.
func RunServerHTTP(
	ctx context.Context,
	serverURL string,
	databaseDSN string,
	databaseMigrationsDir string,
	jwtSecretKey string,
	jwtExp time.Duration,
) error {
	// 1. Подключение к базе данных
	log.Println("инициализация базы данных...")
	conn, err := db.New("sqlite", databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 2. Выполнение миграций
	log.Println("выполнение миграций...")
	goose.SetDialect("sqlite")
	if databaseMigrationsDir != "" {
		if err := goose.Up(conn.DB, databaseMigrationsDir); err != nil {
			return err
		}
	}

	// 3. JWT
	log.Println("инициализация JWT...")
	jwt := jwt.New(
		jwt.WithSecret(jwtSecretKey),
		jwt.WithTTL(jwtExp),
	)

	// 4. Менеджер транзакций
	tx := tx.New(conn)

	// 5. Репозитории
	log.Println("инициализация репозиториев...")
	userReaderRepository := repositories.NewUserReaderRepository(conn)
	userWriterRepository := repositories.NewUserWriterRepository(conn)
	deviceReaderRepository := repositories.NewDeviceReaderRepository(conn)
	deviceWriterRepository := repositories.NewDeviceWriterRepository(conn)

	// 6. Сервис аутентификации & handler
	log.Println("инициализация сервисов...")
	authService := services.NewAuthService(
		userReaderRepository,
		userWriterRepository,
		deviceReaderRepository,
		deviceWriterRepository,
		jwt,
		crypto.HashPassword,
		crypto.GenerateRSAKeys,
		crypto.GenerateDEK,
		crypto.EncryptDEK,
		crypto.RSAPrivateKeyToPEM,
	)

	// 7. Обработчики
	log.Println("инициализация обработчиков...")
	authHandler := handlers.NewAuthHTTPHandler(
		authService,
		validators.ValidateUsername,
		validators.ValidatePassword,
	)

	// 8. Middleware для транзакций
	log.Println("инициализация middleware для транзакций...")
	txMiddleware := middlewares.NewTxMiddleware(tx)

	// 9. Маршрутизатор
	log.Println("инициализация маршрутизатора...")
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(txMiddleware) // <--- добавлен middleware для транзакций
	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
	})

	// 10. HTTP сервер
	log.Println("инициализация сервера...")
	srv := &http.Server{
		Addr:    serverURL,
		Handler: router,
	}

	// 11. Обработка сигналов завершения
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		log.Println("запуск HTTP сервера")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	<-ctx.Done()
	log.Println("получен сигнал завершения")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("завершение работы HTTP сервера...")
	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		errCh <- err
	}

	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	default:
		return nil
	}
}
