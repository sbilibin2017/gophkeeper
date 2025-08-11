package login

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/transport/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
)

// RunHTTP performs a login request using HTTP protocol.
func RunHTTP(
	ctx context.Context,
	serverURL, username, password string,
) (string, error) {
	client := http.New(serverURL)
	authFacade := facades.NewAuthHTTPFacade(client)
	return authFacade.Login(ctx, username, password)
}

// RunGRPC performs a login request using gRPC protocol.
func RunGRPC(
	ctx context.Context,
	serverURL, username, password string,
) (string, error) {
	conn, err := grpc.New(serverURL)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	authFacade := facades.NewAuthGRPCFacade(conn)
	return authFacade.Login(ctx, username, password)
}
