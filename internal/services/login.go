package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
)

// LoginHTTP отправляет HTTP-запрос для логина пользователя.
// Принимает контекст, HTTP клиент resty.Client, имя пользователя и пароль.
// Возвращает JWT токен и ошибку, если регистрация не удалась.
func LoginHTTP(
	ctx context.Context,
	client *resty.Client,
	username,
	password string,
) (string, error) {
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	var result struct {
		Token string `json:"token"`
	}

	resp, err := client.R().
		SetContext(ctx).
		SetBody(payload).
		SetResult(&result).
		Post("/login")

	if err != nil {
		return "", fmt.Errorf("failed to send login request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return "", fmt.Errorf("login failed: status %d, response: %s", resp.StatusCode(), resp.String())
	}

	return result.Token, nil
}

// LoginGRPC отправляет GRPC-запрос для логина пользователя.
// Принимает контекст, gRPC клиент RegisterServiceClient, имя пользователя и пароль.
// Возвращает ошибку, если регистрация не удалась.
func LoginGRPC(
	ctx context.Context,
	client pb.LoginServiceClient,
	username,
	password string,
) (string, error) {
	req := &pb.Credentials{
		Username: username,
		Password: password,
	}

	resp, err := client.Login(ctx, req)
	if err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", fmt.Errorf("login error: %s", resp.Error)
	}

	return resp.Token, nil
}
