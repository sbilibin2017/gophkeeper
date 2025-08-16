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
func (h *SecretHTTPFacade) Save(
	ctx context.Context,
	secretID, userID, secretName, secretType string,
	encryptedPayload, nonce []byte,
	meta string,
) error {
	body := map[string]any{
		"secret_id":         secretID,
		"user_id":           userID,
		"secret_name":       secretName,
		"secret_type":       secretType,
		"encrypted_payload": encryptedPayload,
		"nonce":             nonce,
		"meta":              meta,
	}

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(userID).
		SetBody(body).
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
func (h *SecretHTTPFacade) Get(
	ctx context.Context,
	userID, secretName string,
) (*models.SecretDB, error) {
	var secret models.SecretDB

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(userID).
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
func (h *SecretHTTPFacade) List(
	ctx context.Context,
	userID string,
) ([]*models.SecretDB, error) {
	var secrets []*models.SecretDB

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(userID).
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
