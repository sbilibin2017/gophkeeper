package login

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/http"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

// RunLoginGRPC performs user login via gRPC.
func RunLoginGRPC(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, username, password string) (*models.AuthResponse, error) {
	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	if err := createLoginTables(ctx, dbConn); err != nil {
		return nil, err
	}

	conn, err := grpc.New(
		strings.TrimPrefix(authURL, "grpc://"),
		grpc.WithTLSCert(grpc.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    2 * time.Second,
			MaxWait: 10 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	facade := auth.NewLoginGRPCFacade(client)

	return facade.Login(ctx, &models.AuthRequest{
		Username: username,
		Password: password,
	})
}

// RunLoginHTTP performs user login via HTTP.
func RunLoginHTTP(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, username, password string) (*models.AuthResponse, error) {
	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	if err := createLoginTables(ctx, dbConn); err != nil {
		return nil, err
	}

	client, err := http.New(
		authURL,
		http.WithTLSCert(http.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count: 3,
			Wait:  2 * time.Second,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	facade := auth.NewLoginHTTPFacade(client)

	return facade.Login(ctx, &models.AuthRequest{
		Username: username,
		Password: password,
	})
}

func createLoginTables(ctx context.Context, dbConn *sqlx.DB) error {
	if err := bankcard.CreateClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := text.CreateClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := binary.CreateClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := user.CreateClientTable(ctx, dbConn); err != nil {
		return err
	}
	return nil
}
