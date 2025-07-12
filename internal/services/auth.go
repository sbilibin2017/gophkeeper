package services

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// AuthHTTP выполняет аутентификацию через HTTP с помощью resty клиента.
func AuthHTTP(
	ctx context.Context,
	client *resty.Client,
	secret *models.UsernamePassword,
) (string, error) {
	var result struct {
		Token string `json:"token"`
	}

	resp, err := client.R().
		SetContext(ctx).
		SetBody(secret).
		SetResult(&result).
		Post("/auth") // здесь POST на /auth, как в протоколе

	if err != nil {
		return "", fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	if resp.IsError() {
		return "", fmt.Errorf("server returned an error response: %s", resp.Status())
	}

	if result.Token == "" {
		return "", fmt.Errorf("token not received in server response")
	}

	return result.Token, nil
}

// AuthGRPC выполняет аутентификацию через gRPC клиента.
func AuthGRPC(
	ctx context.Context,
	client pb.AuthServiceClient, // используем AuthServiceClient
	secret *models.UsernamePassword,
) (string, error) {
	req := &pb.AuthRequest{ // AuthRequest, не RegisterRequest
		Username: secret.Username,
		Password: secret.Password,
	}

	resp, err := client.Auth(ctx, req) // вызываем Auth метод
	if err != nil {
		return "", fmt.Errorf("failed to perform gRPC request: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("authentication error: %s", resp.Error)
	}

	if resp.Token == "" {
		return "", fmt.Errorf("token not received from gRPC server")
	}

	return resp.Token, nil
}
