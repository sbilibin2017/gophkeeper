package register

import (
	"context"

	"github.com/sbilibin2017/gophkeeper/internal/facades"
	"github.com/sbilibin2017/gophkeeper/internal/rsa"
	"github.com/sbilibin2017/gophkeeper/internal/transport/grpc"
	"github.com/sbilibin2017/gophkeeper/internal/transport/http"
)

func RunHTTP(ctx context.Context, serverURL, username, password string) (string, error) {
	pubPEM, privPEM, err := rsa.GenerateRSAKeys(username)
	if err != nil {
		return "", err
	}

	if err := rsa.SaveKeyPair(username, pubPEM, privPEM); err != nil {
		return "", err
	}

	client := http.New(serverURL)
	authFacade := facades.NewAuthHTTPFacade(client)

	token, err := authFacade.Register(ctx, username, password)
	if err != nil {
		return "", err
	}

	return token, nil
}

func RunGRPC(ctx context.Context, serverURL, username, password string) (string, error) {
	pubPEM, privPEM, err := rsa.GenerateRSAKeys(username)
	if err != nil {
		return "", err
	}

	if err := rsa.SaveKeyPair(username, pubPEM, privPEM); err != nil {
		return "", err
	}

	conn, err := grpc.New(serverURL)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	authFacade := facades.NewAuthGRPCFacade(conn)

	token, err := authFacade.Register(ctx, username, password)
	if err != nil {
		return "", err
	}

	return token, nil
}
