package server

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

func RunHTTP(
	ctx context.Context,
	serverURL string,
	databaseDSN string,
	databaseMigrationsDir string,
	jwtSecretKey string,
	jwtExp time.Duration,
) error {
	// 1. Connect to database
	log.Println("initializing db...")
	conn, err := db.New("sqlite", databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 2. Run migrations
	log.Println("running migrations...")
	goose.SetDialect("sqlite")
	if databaseMigrationsDir != "" {
		if err := goose.Up(conn.DB, databaseMigrationsDir); err != nil {
			return err
		}
	}

	// 3. JWT
	log.Println("initializing jwt...")
	jwt := jwt.New(
		jwt.WithSecret(jwtSecretKey),
		jwt.WithTTL(jwtExp),
	)

	// 4. Transaction manager
	tx := tx.New(conn)

	// 5. Repositories
	log.Println("initializing repositories...")
	userReaderRepository := repositories.NewUserReaderRepository(conn)
	userWriterRepository := repositories.NewUserWriterRepository(conn)
	deviceReaderRepository := repositories.NewDeviceReaderRepository(conn)
	deviceWriterRepository := repositories.NewDeviceWriterRepository(conn)

	// 6. Auth service & handler
	log.Println("initializing services...")
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

	// 7. Handlers
	log.Println("initializing handlers...")
	authHandler := handlers.NewAuthHTTPHandler(
		authService,
		validators.ValidateUsername,
		validators.ValidatePassword,
	)

	// 8. Tx middleware
	log.Println("initializing transaction middleware...")
	txMiddleware := middlewares.NewTxMiddleware(tx)

	// 9. Router
	log.Println("initializing router...")
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(txMiddleware) // <--- добавили middleware для транзакций
	router.Route("/api/v1", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
	})

	// 10. HTTP server
	log.Println("initializing server...")
	srv := &http.Server{
		Addr:    serverURL,
		Handler: router,
	}

	// 11. Handle shutdown signals
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		log.Println("starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Println("shutting down HTTP server...")
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
