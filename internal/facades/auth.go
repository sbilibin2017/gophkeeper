package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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
	req *models.RegisterRequest,
) (resp *models.RegisterResponse, token string, err error) {

	resp = &models.RegisterResponse{}

	httpResp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(resp).
		Post("/register")
	if err != nil {
		return nil, "", err
	}

	if httpResp.IsError() {
		return nil, "", fmt.Errorf("http error: %s", httpResp.Status())
	}

	token, err = h.tokenGetter.GetFromResponse(httpResp)
	if err != nil {
		return nil, "", err
	}

	return resp, token, nil
}

// Login выполняет аутентификацию пользователя на сервере.
func (h *AuthHTTPFacade) Login(
	ctx context.Context,
	req *models.LoginRequest,
) (resp *models.LoginResponse, token string, err error) {

	resp = &models.LoginResponse{}

	httpResp, err := h.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(resp).
		Post("/login")
	if err != nil {
		return nil, "", err
	}

	if httpResp.IsError() {
		return nil, "", fmt.Errorf("http error: %s", httpResp.Status())
	}

	token, err = h.tokenGetter.GetFromResponse(httpResp)
	if err != nil {
		return nil, "", err
	}

	return resp, token, nil
}
