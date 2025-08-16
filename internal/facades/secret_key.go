package facades

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretKeyHTTPFacade предоставляет методы для работы с секретами пользователя через HTTP API.
type SecretKeyHTTPFacade struct {
	client *resty.Client
}

// NewSecretKeyHTTPFacade создаёт новый экземпляр SecretKeyHTTPFacade с указанным HTTP клиентом.
func NewSecretKeyHTTPFacade(client *resty.Client) *SecretKeyHTTPFacade {
	return &SecretKeyHTTPFacade{client: client}
}

// Get возвращает секрет по его ID.
func (h *SecretKeyHTTPFacade) Get(
	ctx context.Context,
	token string,
	secretID string,
) (*models.SecretKeyResponse, error) {
	var secret models.SecretKeyResponse

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&secret).
		Get(fmt.Sprintf("/get/%s", secretID))
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error: %s", resp.Status())
	}

	return &secret, nil
}
