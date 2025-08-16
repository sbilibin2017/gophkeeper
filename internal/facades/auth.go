package facades

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// TokenGetter интерфейс для получения токена из HTTP-ответа.
type TokenGetter interface {
	// GetFromResponse извлекает токен из ответа сервера.
	GetFromResponse(resp *resty.Response) (string, error)
}

// AuthHTTPFacade предоставляет методы для взаимодействия с сервером аутентификации через HTTP.
type AuthHTTPFacade struct {
	client      *resty.Client
	tokenGetter TokenGetter
}

// NewAuthHTTPFacade создаёт новый фасад аутентификации с указанным HTTP клиентом и TokenGetter.
func NewAuthHTTPFacade(
	client *resty.Client,
	tokenGetter TokenGetter,
) *AuthHTTPFacade {
	return &AuthHTTPFacade{
		client:      client,
		tokenGetter: tokenGetter,
	}
}

// Register регистрирует нового пользователя на сервере.
func (h *AuthHTTPFacade) Register(
	ctx context.Context,
	username string,
	password string,
) (userID, deviceID, privateKey, token string, err error) {

	req := map[string]string{
		"username": username,
		"password": password,
	}

	httpResp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/register")
	if err != nil {
		return "", "", "", "", err
	}

	if httpResp.IsError() {
		return "", "", "", "", fmt.Errorf("http error: %s", httpResp.Status())
	}

	// Распарсить тело ответа
	var result map[string]string
	if err := json.Unmarshal(httpResp.Body(), &result); err != nil {
		return "", "", "", "", err
	}

	userID = result["user_id"]
	deviceID = result["device_id"]
	privateKey = result["private_key"]

	token, err = h.tokenGetter.GetFromResponse(httpResp)
	if err != nil {
		return "", "", "", "", err
	}

	return userID, deviceID, privateKey, token, nil
}

// Login выполняет аутентификацию пользователя на сервере.
func (h *AuthHTTPFacade) Login(
	ctx context.Context,
	username string,
	password string,
) (token string, err error) {

	req := map[string]string{
		"username": username,
		"password": password,
	}

	httpResp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		Post("/login")
	if err != nil {
		return "", err
	}

	if httpResp.IsError() {
		return "", fmt.Errorf("http error: %s", httpResp.Status())
	}

	token, err = h.tokenGetter.GetFromResponse(httpResp)
	if err != nil {
		return "", err
	}

	return token, nil
}
