package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RegisterUserHTTP выполняет регистрацию пользователя через HTTP API.
// Возвращает токен или ошибку.
func RegisterUserHTTP(ctx context.Context, client *resty.Client, req *models.AuthRequest) (string, error) {
	var respBody models.AuthResponse

	resp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&respBody).
		Post("/register")
	if err != nil {
		return "", fmt.Errorf("server unavailable: %w", err)
	}

	if resp.IsError() {
		return "", errors.New("registration error: " + resp.Status())
	}

	if respBody.Token == "" {
		return "", errors.New("token not received from server")
	}

	return respBody.Token, nil
}

// RegisterUserGRPC выполняет регистрацию пользователя через gRPC API.
// Возвращает токен или ошибку.
func RegisterUserGRPC(ctx context.Context, client pb.AuthServiceClient, req *models.AuthRequest) (string, error) {
	resp, err := client.Register(ctx, &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return "", err
	}

	if resp.Token == "" {
		return "", errors.New("token not received from server")
	}

	return resp.Token, nil
}

// LoginUserHTTP выполняет аутентификацию пользователя через HTTP API.
// Возвращает токен или ошибку.
func LoginUserHTTP(ctx context.Context, client *resty.Client, req *models.AuthRequest) (string, error) {
	var respBody models.AuthResponse

	resp, err := client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&respBody).
		Post("/login")
	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", errors.New("authentication error: " + resp.Status())
	}

	if respBody.Token == "" {
		return "", errors.New("token not received from server")
	}

	return respBody.Token, nil
}

// LoginUserGRPC выполняет аутентификацию пользователя через gRPC API.
// Возвращает токен или ошибку.
func LoginUserGRPC(ctx context.Context, client pb.AuthServiceClient, req *models.AuthRequest) (string, error) {
	resp, err := client.Login(ctx, &pb.AuthRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return "", err
	}

	if resp.Token == "" {
		return "", errors.New("token not received from server")
	}

	return resp.Token, nil
}
