package logout

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
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

// RunLogoutGRPC performs user logout via gRPC.
func RunLogoutGRPC(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, token string) error {
	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	conn, err := grpc.New(
		strings.TrimPrefix(authURL, "grpc://"),
		grpc.WithTLSCert(grpc.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
		grpc.WithAuthToken(token),
		grpc.WithRetryPolicy(grpc.RetryPolicy{
			Count:   3,
			Wait:    2 * time.Second,
			MaxWait: 10 * time.Second,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection: %w", err)
	}
	defer conn.Close()

	client := pb.NewAuthServiceClient(conn)
	facade := auth.NewLogoutGRPCFacade(client)

	if err := facade.Logout(ctx); err != nil {
		return err
	}

	return dropClientTables(ctx, dbConn)
}

// RunLogoutHTTP performs user logout via HTTP.
func RunLogoutHTTP(ctx context.Context, authURL, tlsCertFile, tlsKeyFile, token string) error {
	dbConn, err := db.NewDB("sqlite", "client.db")
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}
	defer dbConn.Close()

	client, err := http.New(
		authURL,
		http.WithTLSCert(http.TLSCert{CertFile: tlsCertFile, KeyFile: tlsKeyFile}),
		http.WithRetryPolicy(http.RetryPolicy{
			Count: 3,
			Wait:  2 * time.Second,
		}),
		http.WithAuthToken(token),
	)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}

	facade := auth.NewLogoutHTTPFacade(client)

	if err := facade.Logout(ctx); err != nil {
		return err
	}

	return dropClientTables(ctx, dbConn)
}

// helper to initialize DB and drop client tables
func dropClientTables(ctx context.Context, dbConn *sqlx.DB) error {
	if err := bankcard.DropClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := text.DropClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := binary.DropClientTable(ctx, dbConn); err != nil {
		return err
	}
	if err := user.DropClientTable(ctx, dbConn); err != nil {
		return err
	}
	return nil
}
