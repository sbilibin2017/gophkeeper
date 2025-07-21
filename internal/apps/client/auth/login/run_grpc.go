package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/configs/clients/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/configs/db"
	"github.com/sbilibin2017/gophkeeper/internal/facades/auth"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/bankcard"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/binary"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/text"
	"github.com/sbilibin2017/gophkeeper/internal/repositories/user"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc/auth"
)

// NewRunGRPC returns a closure function that performs user authentication
// via gRPC using the provided server address and TLS certificate files.
//
// Parameters:
//   - authURL: The full gRPC service address (e.g., "grpc://localhost:50051").
//   - tlsCertFile: Path to the TLS certificate file for the client.
//   - tlsKeyFile: Path to the TLS key file for the client.
//
// The returned function accepts:
//   - ctx: Context for timeout and cancellation control.
//   - username: The username of the user attempting to authenticate.
//   - password: The user's password.
//
// Returns:
//   - *models.AuthResponse: Contains access and refresh tokens if authentication succeeds.
//   - error: Any error encountered during DB setup, connection, or login process.
func NewRunGRPC(authURL, tlsCertFile, tlsKeyFile string) func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
	return func(ctx context.Context, username, password string) (*models.AuthResponse, error) {
		dbConn, err := db.NewDB("sqlite", "client.db")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to DB: %w", err)
		}
		defer dbConn.Close()

		if err := bankcard.CreateClientTable(ctx, dbConn); err != nil {
			return nil, fmt.Errorf("failed to create bankcard table: %w", err)
		}
		if err := text.CreateClientTable(ctx, dbConn); err != nil {
			return nil, fmt.Errorf("failed to create text table: %w", err)
		}
		if err := binary.CreateClientTable(ctx, dbConn); err != nil {
			return nil, fmt.Errorf("failed to create binary table: %w", err)
		}
		if err := user.CreateClientTable(ctx, dbConn); err != nil {
			return nil, fmt.Errorf("failed to create user table: %w", err)
		}

		grpcConn, err := grpc.New(
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
		defer grpcConn.Close()

		grpcClient := pb.NewAuthServiceClient(grpcConn)
		facade := auth.NewLoginGRPCFacade(grpcClient)

		authReq := &models.AuthRequest{
			Username: username,
			Password: password,
		}

		resp, err := facade.Login(ctx, authReq)
		if err != nil {
			return nil, err
		}

		return resp, nil
	}
}
