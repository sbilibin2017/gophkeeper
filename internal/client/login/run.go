package login

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/transport/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
)

// RunHTTP performs a login request using HTTP protocol.
func RunHTTP(
	ctx context.Context,
	serverURL string,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	client := http.New(serverURL)
	authFacade := facades.NewAuthHTTPFacade(client)
	return authFacade.Login(ctx, req)
}

// RunGRPC performs a login request using gRPC protocol.
func RunGRPC(
	ctx context.Context,
	serverURL string,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	conn, err := grpc.New(serverURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	authFacade := facades.NewAuthGRPCFacade(conn)
	return authFacade.Login(ctx, req)
}
