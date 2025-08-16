package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretHTTPFacade предоставляет методы для работы с секретами пользователя через HTTP API.
type SecretHTTPFacade struct {
	client *resty.Client
}

func NewSecretHTTPFacade(client *resty.Client) *SecretHTTPFacade {
	return &SecretHTTPFacade{client: client}
}

// Save вставляет или обновляет секрет на сервере
func (h *SecretHTTPFacade) Save(ctx context.Context, token string, req *models.SecretRequest) error {
	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetBody(req).
		Post("/save")
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("http error: %s", resp.Status())
	}

	return nil
}

// Get возвращает секрет по secretName
func (h *SecretHTTPFacade) Get(ctx context.Context, token, secretName string) (*models.SecretResponse, error) {
	var secret models.SecretResponse

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(fmt.Sprintf("/get/%s", secretName))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error: %s", resp.Status())
	}

	return &secret, nil
}

// List возвращает все секреты пользователя
func (h *SecretHTTPFacade) List(ctx context.Context, token string) ([]*models.SecretResponse, error) {
	var secrets []*models.SecretResponse

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secrets).
		Get("/list")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error: %s", resp.Status())
	}

	return secrets, nil
}
