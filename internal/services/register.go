package services

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// RegisterHTTP выполняет регистрацию через HTTP с помощью resty клиента.
func RegisterHTTP(
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
		Post("/register")

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

// RegisterGRPC выполняет регистрацию через gRPC клиента.
func RegisterGRPC(
	ctx context.Context,
	client pb.RegisterServiceClient,
	secret *models.UsernamePassword,
) (string, error) {
	req := &pb.RegisterRequest{
		Username: secret.Username,
		Password: secret.Password,
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to perform gRPC request: %w", err)
	}

	if resp.Error != "" {
		return "", fmt.Errorf("registration error: %s", resp.Error)
	}

	if resp.Token == "" {
		return "", fmt.Errorf("token not received from gRPC server")
	}

	return resp.Token, nil
}
