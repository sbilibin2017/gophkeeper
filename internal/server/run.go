package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/pressly/goose/v3"
	"github.com/sbilibin2017/gophkeeper/internal/db"
	"github.com/sbilibin2017/gophkeeper/internal/handlers"
	"github.com/sbilibin2017/gophkeeper/internal/hasher"
	"github.com/sbilibin2017/gophkeeper/internal/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/sbilibin2017/gophkeeper/internal/validators"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	grpcConn "google.golang.org/grpc"
)

// RunHTTP initializes and runs the HTTP server with authentication routes.
// It sets up the database, runs migrations, creates the authentication service,
// configures routes, and starts listening on the specified address.
// The server gracefully shuts down when the context is canceled.
func RunHTTP(
	ctx context.Context,
	apiVersion string,
	databaseDriver string,
	databaseDSN string,
	databaseMaxOpenConns int,
	databaseMaxIdleConns int,
	databaseConnMaxLifetime time.Duration,
	jwtSecretKey string,
	jwtExp time.Duration,
	pathToMigrationsDir string,
	serverAddr string,
) error {
	db, err := db.New(
		databaseDriver,
		databaseDSN,
		db.WithMaxOpenConns(databaseMaxOpenConns),
		db.WithMaxIdleConns(databaseMaxIdleConns),
		db.WithConnMaxLifetime(databaseConnMaxLifetime),
	)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}

	if err := goose.Up(db.DB, pathToMigrationsDir); err != nil {
		return err
	}

	userReadRepo := repositories.NewUserReadRepository(db)
	userWriteRepo := repositories.NewUserWriteRepository(db)

	hasher := hasher.New()
	jwtManager := jwt.New(
		jwt.WithSecret(jwtSecretKey),
		jwt.WithLifetime(jwtExp),
	)

	authService := services.NewAuthService(userReadRepo, userWriteRepo, hasher, jwtManager)

	httpHandler := handlers.NewHTTPHandler(
		authService,
		validators.ValidateUsername,
		validators.ValidatePassword,
	)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post(apiVersion+"/register", httpHandler.Register)
	r.Post(apiVersion+"/login", httpHandler.Login)

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	errCh := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}

// RunGRPC initializes and runs the gRPC server for authentication services.
// It sets up the database, runs migrations, creates the authentication service,
// registers the gRPC handler, and starts listening on the specified address.
// The server gracefully stops when the context is canceled.
func RunGRPC(
	ctx context.Context,
	databaseDriver string,
	databaseDSN string,
	databaseMaxOpenConns int,
	databaseMaxIdleConns int,
	databaseConnMaxLifetime time.Duration,
	jwtSecretKey string,
	jwtExp time.Duration,
	pathToMigrationsDir string,
	serverAddr string,
	opts ...grpcConn.ServerOption,
) error {
	db, err := db.New(
		databaseDriver,
		databaseDSN,
		db.WithMaxOpenConns(databaseMaxOpenConns),
		db.WithMaxIdleConns(databaseMaxIdleConns),
		db.WithConnMaxLifetime(databaseConnMaxLifetime),
	)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := goose.SetDialect("sqlite"); err != nil {
		return err
	}

	if err := goose.Up(db.DB, pathToMigrationsDir); err != nil {
		return err
	}

	userReadRepo := repositories.NewUserReadRepository(db)
	userWriteRepo := repositories.NewUserWriteRepository(db)

	hasher := hasher.New()
	jwtManager := jwt.New(
		jwt.WithSecret(jwtSecretKey),
		jwt.WithLifetime(jwtExp),
	)

	authService := services.NewAuthService(userReadRepo, userWriteRepo, hasher, jwtManager)

	grpcServer := grpcConn.NewServer(opts...)

	grpcHandler := handlers.NewGRPCHandler(
		authService,
		validators.ValidateUsername,
		validators.ValidatePassword,
	)

	pb.RegisterAuthServiceServer(grpcServer, grpcHandler)

	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	done := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(done)
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	case <-done:
	}

	return nil
}
