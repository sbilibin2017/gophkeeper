package client

import (
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretKeyHTTPClient предоставляет методы для работы с секретами пользователя через HTTP API.
type SecretKeyHTTPClient struct {
	client *resty.Client
}

// NewSecretKeyHTTPClient создаёт новый экземпляр SecretKeyHTTPClient с указанным HTTP клиентом.
func NewSecretKeyHTTPClient(client *resty.Client) *SecretKeyHTTPClient {
	return &SecretKeyHTTPClient{client: client}
}

// Get возвращает секрет по его ID.
func (h *SecretKeyHTTPClient) Get(
	ctx context.Context,
	token string,
	secretID string,
) (*models.SecretKeyResponse, error) {
	var respData models.SecretKeyResponse

	httpResp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&respData).
		Get(fmt.Sprintf("/get/%s", secretID))
	if err != nil {
		return nil, err
	}

	if httpResp.IsError() {
		return nil, fmt.Errorf("http error: %s", httpResp.Status())
	}

	// Возвращаем новый объект с декодированным ключом
	secretKey := &models.SecretKeyResponse{
		SecretKeyID:     respData.SecretKeyID,
		SecretID:        respData.SecretID,
		DeviceID:        respData.DeviceID,
		EncryptedAESKey: respData.EncryptedAESKey,
		CreatedAt:       respData.CreatedAt,
		UpdatedAt:       respData.UpdatedAt,
	}

	return secretKey, nil
}
