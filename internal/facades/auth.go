package facades

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
)

// AuthHTTPFacade фасад для работы с HTTP-сервисом авторизации
type AuthHTTPFacade struct {
	client      *resty.Client                              // HTTP клиент Resty
	tokenGetter func(resp *resty.Response) (string, error) // Функция для извлечения токена из ответа
}

// NewAuthHTTPFacade создает новый AuthHTTPFacade
func NewAuthHTTPFacade(
	client *resty.Client,
	tokenGetter func(resp *resty.Response) (string, error),
) *AuthHTTPFacade {
	return &AuthHTTPFacade{
		client:      client,
		tokenGetter: tokenGetter,
	}
}

// Register регистрирует пользователя и устройство через HTTP и возвращает приватный ключ и токен
func (f *AuthHTTPFacade) Register(
	ctx context.Context,
	username string,
	password string,
	deviceID string,
) ([]byte, string, error) {
	reqBody := map[string]string{
		"username":  username,
		"password":  password,
		"device_id": deviceID,
	}

	resp, err := f.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		Post("/register")
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, "", fmt.Errorf("registration failed with status %d: %s", resp.StatusCode(), resp.String())
	}

	privKey := resp.Body()
	token, err := f.tokenGetter(resp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract token: %w", err)
	}

	return privKey, token, nil
}
