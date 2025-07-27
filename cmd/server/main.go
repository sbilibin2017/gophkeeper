package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/pressly/goose"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/configs/scheme"
	"github.com/sbilibin2017/gophkeeper/internal/jwt"
	"github.com/sbilibin2017/gophkeeper/internal/repositories"
)

func main() {
	flag.Parse()
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

var (
	serverURL    string
	databaseDSN  string
	JWTSecretKey string
	JWTExp       time.Duration
)

func init() {
	flag.StringVar(&serverURL, "server-url", "", "Server URL")
	flag.StringVar(&databaseDSN, "database-dsn", "", "Database DSN (Data Source Name)")
	flag.StringVar(&JWTSecretKey, "jwt-secret-key", "", "JWT secret key")
	flag.DurationVar(&JWTExp, "jwt-exp", 0, "JWT expiration duration (e.g. 24h, 30m)")
}

func run(ctx context.Context) error {
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

	if err := goose.Up(dbConn.DB, "../../../migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	writeRepo := repositories.NewSecretWriteRepository(dbConn)
	readRepo := repositories.NewSecretReadRepository(dbConn)

	jwtManager := jwt.New(
		jwt.WithSecret(JWTSecretKey),
		jwt.WithLifetime(JWTExp),
	)

	schm := scheme.GetSchemeFromURL(serverURL)

	switch schm {
	case scheme.HTTP, scheme.HTTPS:
		client, err := http.New(serverURL, http.WithRetryPolicy(http.RetryPolicy{
			Count:   3,
			Wait:    500 * time.Millisecond,
			MaxWait: 2 * time.Second,
		}))
		if err != nil {
			return err
		}
		_ = client

	case scheme.GRPC:
		conn, err := grpc.New(serverURL, grpc.WithRetryPolicy(
			grpc.RetryPolicy{
				Count:   3,
				Wait:    500 * time.Millisecond,
				MaxWait: 2 * time.Second,
			}))
		if err != nil {
			return err
		}
		defer conn.Close()
		_ = conn

	default:
		return fmt.Errorf("unsupported scheme: %s", schm)
	}

	return nil
}
