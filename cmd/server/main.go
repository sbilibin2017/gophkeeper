// Package main implements the entry point for the GophKeeper server application.
// It supports HTTP(S) and gRPC protocols with JWT authentication, database migrations,
// and graceful shutdown on system signals.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/pressly/goose"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sbilibin2017/gophkeeper/internal/db"
	"github.com/sbilibin2017/gophkeeper/internal/grpcservers"
	"github.com/sbilibin2017/gophkeeper/internal/handlers"
	"github.com/sbilibin2017/gophkeeper/internal/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/scheme"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

// main parses the command line flags and runs the server with context.
func main() {
	printBuildInfo()

	flag.Parse()

	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

var (
	// serverURL holds the URL where the server will run.
	serverURL string

	// databaseDSN contains the database connection string.
	databaseDSN string

	// jwtSecretKey is the secret key used for signing JWT tokens.
	jwtSecretKey string

	// jwtExp defines the expiration duration for JWT tokens.
	jwtExp time.Duration
)

// init registers command line flags for configuring server parameters.
func init() {
	flag.StringVar(&serverURL, "server-url", "", "Server URL (e.g. http://localhost:8080 or localhost:8080)")
	flag.StringVar(&databaseDSN, "database-dsn", "", "Database DSN (Data Source Name)")
	flag.StringVar(&jwtSecretKey, "jwt-secret-key", "", "JWT secret key")
	flag.DurationVar(&jwtExp, "jwt-exp", 0, "JWT expiration duration (e.g. 24h, 30m)")
}

// printBuildInfo outputs the build version and build date to standard output.
func printBuildInfo() {
	fmt.Printf("GophKeeper Server - Build version: %s, Build date: %s\n", buildVersion, buildDate)
}

// run initializes the database connection, runs migrations, sets up repositories and JWT,
// then starts either an HTTP(S) or gRPC server depending on the serverURL scheme.
// It listens for system signals for graceful shutdown.
func run(ctx context.Context) error {
	// Connect to the database with connection pooling settings.
	dbConn, err := db.New(
		"sqlite",
		databaseDSN,
		db.WithMaxOpenConns(10),
		db.WithMaxIdleConns(5),
		db.WithConnMaxLifetime(30*time.Minute),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	// Set goose dialect and apply migrations.
	if err := goose.SetDialect("sqlite"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(dbConn.DB, "../../../migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize repositories for users and secrets.
	userWriteRepo := repositories.NewUserWriteRepository(dbConn)
	userReadRepo := repositories.NewUserReadRepository(dbConn)
	secretWriter := repositories.NewSecretWriteRepository(dbConn)
	secretReader := repositories.NewSecretReadRepository(dbConn)

	// Create JWT manager with secret and expiration.
	jwtManager := jwt.New(
		jwt.WithSecret(jwtSecretKey),
		jwt.WithLifetime(jwtExp),
	)

	// Determine server scheme from the server URL.
	schm := scheme.GetSchemeFromURL(serverURL)

	// Parse serverURL to extract address for listeners.
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("failed to parse server-url: %w", err)
	}

	addr := parsedURL.Host
	if addr == "" {
		// If the input is just host:port without scheme, parsedURL.Host is empty, Path has value.
		addr = serverURL
	}

	// Create a cancellable context that listens for shutdown signals.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	switch schm {
	case scheme.HTTP, scheme.HTTPS:
		// Setup HTTP router with logging and recovery middleware.
		r := chi.NewRouter()
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		// Register authentication endpoints.
		r.Post("/api/v1/register", handlers.NewRegisterHandler(userReadRepo, userWriteRepo, jwtManager))
		r.Post("/api/v1/login", handlers.NewLoginHandler(userReadRepo, jwtManager))

		// Register secret management endpoints.
		r.Post("/api/v1/secrets", handlers.NewSecretAddHandler(secretWriter, jwtManager))
		r.Get("/api/v1/secrets/{secret_type}/{secret_name}", handlers.NewSecretGetHandler(secretReader, jwtManager))
		r.Get("/api/v1/secrets", handlers.NewSecretListHandler(secretReader, jwtManager))

		srv := &http.Server{
			Addr:    addr,
			Handler: r,
		}

		// Start the HTTP server in a background goroutine.
		serverErrors := make(chan error, 1)
		go func() {
			log.Printf("Starting HTTP server at %s\n", addr)
			serverErrors <- srv.ListenAndServe()
		}()

		// Wait for server error or shutdown signal.
		select {
		case err := <-serverErrors:
			return err
		case <-ctx.Done():
			log.Println("Shutdown signal received for HTTP server")

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := srv.Shutdown(shutdownCtx); err != nil {
				log.Printf("HTTP server Shutdown error: %v", err)
				return err
			}

			log.Println("HTTP server stopped gracefully")
		}

	case scheme.GRPC:
		// Setup gRPC server and register services.
		grpcServer := grpc.NewServer()

		authServer := grpcservers.NewAuthServer(userWriteRepo, userReadRepo, jwtManager)
		pb.RegisterAuthServiceServer(grpcServer, authServer)

		secretWriteServer := grpcservers.NewSecretWriteServer(secretWriter, jwtManager)
		pb.RegisterSecretWriteServiceServer(grpcServer, secretWriteServer)

		secretReadServer := grpcservers.NewSecretReadServer(secretReader, jwtManager)
		pb.RegisterSecretReadServiceServer(grpcServer, secretReadServer)

		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}

		serverErrors := make(chan error, 1)
		go func() {
			log.Printf("Starting gRPC server at %s\n", addr)
			serverErrors <- grpcServer.Serve(lis)
		}()

		// Wait for server error or shutdown signal.
		select {
		case err := <-serverErrors:
			return err
		case <-ctx.Done():
			log.Println("Shutdown signal received for gRPC server")

			stopped := make(chan struct{})
			go func() {
				grpcServer.GracefulStop()
				close(stopped)
			}()

			select {
			case <-stopped:
				log.Println("gRPC server stopped gracefully")
			case <-time.After(10 * time.Second):
				log.Println("gRPC graceful stop timed out, forcing stop")
				grpcServer.Stop()
			}
		}

	default:
		return fmt.Errorf("unsupported scheme: %s", schm)
	}

	return nil
}
