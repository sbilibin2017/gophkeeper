package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pressly/goose/v3"
	"github.com/sbilibin2017/gophkeeper/internal/db"
	grpcHandlers "github.com/sbilibin2017/gophkeeper/internal/handlers/grpc"
	httpHandlers "github.com/sbilibin2017/gophkeeper/internal/handlers/http"
	"github.com/sbilibin2017/gophkeeper/internal/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
	"github.com/sbilibin2017/gophkeeper/internal/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"google.golang.org/grpc"
)

func main() {
	printBuildInfo()

	flag.Parse()

	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
)

var (
	serverURL    string
	databaseDSN  string
	jwtSecretKey string
	jwtExp       time.Duration
)

func init() {
	flag.StringVar(&serverURL, "server-url", "", "Server URL (e.g. http://localhost:8080 or localhost:8080)")
	flag.StringVar(&databaseDSN, "database-dsn", "", "Database DSN (Data Source Name)")
	flag.StringVar(&jwtSecretKey, "jwt-secret-key", "", "JWT secret key")
	flag.DurationVar(&jwtExp, "jwt-exp", 0, "JWT expiration duration (e.g. 24h, 30m)")
}

func printBuildInfo() {
	fmt.Printf("GophKeeper Server - Build version: %s, Build date: %s\n", buildVersion, buildDate)
}

func run(ctx context.Context) error {
	apiVersion := "/api/v1"
	pathToMigrationsDir := "../../../migrations"

	schm := scheme.GetSchemeFromURL(serverURL)

	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("failed to parse server-url: %w", err)
	}

	addr := parsedURL.Host
	if addr == "" {
		addr = serverURL
	}

	switch schm {
	case scheme.HTTP, scheme.HTTPS:
		return runServerHTTP(ctx, addr, databaseDSN, jwtSecretKey, jwtExp, apiVersion, pathToMigrationsDir)
	case scheme.GRPC:
		return runServerGRPC(ctx, addr, databaseDSN, jwtSecretKey, jwtExp, apiVersion, pathToMigrationsDir)
	default:
		return fmt.Errorf("unsupported scheme: %s", schm)
	}
}

// runServerHTTP runs the HTTP server with full setup and graceful shutdown.
func runServerHTTP(
	ctx context.Context,
	serverAddr string,
	databaseDSN string,
	jwtSecretKey string,
	jwtExp time.Duration,
	apiVersion string,
	pathToMigrationsDir string,
) error {
	// Setup DB connection
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

	if err := goose.SetDialect("sqlite"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(dbConn.DB, pathToMigrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Setup repos & services
	userWriteRepo := repositories.NewUserWriteRepository(dbConn)
	userReadRepo := repositories.NewUserReadRepository(dbConn)
	secretWriter := repositories.NewSecretWriteRepository(dbConn)
	secretReader := repositories.NewSecretReadRepository(dbConn)

	authService := services.NewAuthService(userReadRepo, userWriteRepo)
	secretWriteService := services.NewSecretWriteService(secretWriter)
	secretReadService := services.NewSecretReadService(secretReader)

	jwtManager := jwt.New(
		jwt.WithSecret(jwtSecretKey),
		jwt.WithLifetime(jwtExp),
	)

	// Setup router and middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register routes
	r.Post(apiVersion+"/register", httpHandlers.NewRegisterHandler(authService, jwtManager))
	r.Post(apiVersion+"/login", httpHandlers.NewLoginHandler(authService, jwtManager))

	r.Post(apiVersion+"/secrets", httpHandlers.NewSecretAddHandler(secretWriteService, jwtManager))
	r.Get(apiVersion+"/secrets/{secret_type}/{secret_name}", httpHandlers.NewSecretGetHandler(secretReadService, jwtManager))
	r.Get(apiVersion+"/secrets", httpHandlers.NewSecretListHandler(secretReadService, jwtManager))

	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
	}

	// Listen for shutdown signals
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start server
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Starting HTTP server at %s\n", serverAddr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Wait for shutdown or error
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
		return nil
	}
}

// runServerGRPC runs the gRPC server with full setup and graceful shutdown.
func runServerGRPC(
	ctx context.Context,
	serverAddr string,
	databaseDSN string,
	jwtSecretKey string,
	jwtExp time.Duration,
	apiVersion string,
	pathToMigrationsDir string,
) error {
	// Setup DB connection
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

	if err := goose.SetDialect("sqlite"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(dbConn.DB, pathToMigrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Setup repos & services
	userWriteRepo := repositories.NewUserWriteRepository(dbConn)
	userReadRepo := repositories.NewUserReadRepository(dbConn)
	secretWriter := repositories.NewSecretWriteRepository(dbConn)
	secretReader := repositories.NewSecretReadRepository(dbConn)

	authService := services.NewAuthService(userReadRepo, userWriteRepo)
	secretWriteService := services.NewSecretWriteService(secretWriter)
	secretReadService := services.NewSecretReadService(secretReader)

	jwtManager := jwt.New(
		jwt.WithSecret(jwtSecretKey),
		jwt.WithLifetime(jwtExp),
	)

	grpcServer := grpc.NewServer()

	authServer := grpcHandlers.NewAuthServer(authService, jwtManager)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	secretWriteServer := grpcHandlers.NewSecretWriteServer(secretWriteService, jwtManager)
	pb.RegisterSecretWriteServiceServer(grpcServer, secretWriteServer)

	secretReadServer := grpcHandlers.NewSecretReadServer(secretReadService, jwtManager)
	pb.RegisterSecretReadServiceServer(grpcServer, secretReadServer)

	lis, err := net.Listen("tcp", serverAddr+apiVersion)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Starting gRPC server at %s\n", serverAddr)
		serverErrors <- grpcServer.Serve(lis)
	}()

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
		return nil
	}
}
