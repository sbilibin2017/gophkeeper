package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// TokenGetter интерфейс для получения токена из HTTP-ответа.
type TokenGetter interface {
	// GetFromResponse извлекает токен из ответа сервера.
	GetFromResponse(resp *resty.Response) (string, error)
}

// AuthHTTPClient предоставляет методы для взаимодействия с сервером аутентификации через HTTP.
type AuthHTTPClient struct {
	client      *resty.Client
	tokenGetter TokenGetter
}

// NewAuthHTTPClient создаёт новый клиент аутентификации с указанным HTTP клиентом и TokenGetter.
func NewAuthHTTPClient(
	client *resty.Client,
	tokenGetter TokenGetter,
) *AuthHTTPClient {
	return &AuthHTTPClient{
		client:      client,
		tokenGetter: tokenGetter,
	}
}

// Register регистрирует нового пользователя на сервере.
func (h *AuthHTTPClient) Register(
	ctx context.Context,
	req *models.RegisterRequest,
) (*models.RegisterResponse, error) {
	httpResp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/register")
	if err != nil {
		return nil, err
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("http error: %s", httpResp.Status())
	}

	var regResp models.RegisterResponse
	if err := json.Unmarshal(httpResp.Body(), &regResp); err != nil {
		return nil, fmt.Errorf("failed to decode register response: %w", err)
	}

	// Получаем токен и присваиваем его полю Token
	token, err := h.tokenGetter.GetFromResponse(httpResp)
	if err != nil {
		return nil, err
	}
	regResp.Token = token

	return &regResp, nil
}

// Login выполняет аутентификацию пользователя на сервере.
func (h *AuthHTTPClient) Login(
	ctx context.Context,
	req *models.LoginRequest,
) (*models.LoginResponse, error) {
	httpResp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/login")
	if err != nil {
		return nil, err
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("http error: %s", httpResp.Status())
	}

	token, err := h.tokenGetter.GetFromResponse(httpResp)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{Token: token}, nil
}
