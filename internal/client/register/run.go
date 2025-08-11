package register

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/rsa"
	"github.com/sbilibin2017/gophkeeper/internal/transport/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
)

// RunHTTP performs a registration request using HTTP protocol.
func RunHTTP(
	ctx context.Context,
	serverURL string,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	pubPEM, privPEM, err := rsa.GenerateRSAKeys(req.Username)
	if err != nil {
		return nil, err
	}

	if err := rsa.SaveKeyPair(req.Username, pubPEM, privPEM); err != nil {
		return nil, err
	}

	client := http.New(serverURL)
	authFacade := facades.NewAuthHTTPFacade(client)
	return authFacade.Register(ctx, req)
}

// RunGRPC performs a registration request using gRPC protocol.
func RunGRPC(
	ctx context.Context,
	serverURL string,
	req *models.AuthRequest,
) (*models.AuthResponse, error) {
	pubPEM, privPEM, err := rsa.GenerateRSAKeys(req.Username)
	if err != nil {
		return nil, err
	}

	if err := rsa.SaveKeyPair(req.Username, pubPEM, privPEM); err != nil {
		return nil, err
	}

	conn, err := grpc.New(serverURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	authFacade := facades.NewAuthGRPCFacade(conn)
	return authFacade.Register(ctx, req)
}
