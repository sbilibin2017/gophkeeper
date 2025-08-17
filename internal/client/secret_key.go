package facades

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

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
) (*models.SecretKeyDB, error) {
	var respData struct {
		SecretKeyID     string    `json:"secret_key_id"`
		SecretID        string    `json:"secret_id"`
		DeviceID        string    `json:"device_id"`
		EncryptedAESKey string    `json:"encrypted_aes_key"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
	}

	resp, err := h.client.R().
		SetContext(ctx).
		SetAuthToken(token).
		SetResult(&respData).
		Get("/get")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("http error: %s", resp.Status())
	}

	decodedKey, err := base64.StdEncoding.DecodeString(respData.EncryptedAESKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode AES key: %w", err)
	}

	secretKey := &models.SecretKeyDB{
		SecretKeyID:     respData.SecretKeyID,
		SecretID:        respData.SecretID,
		DeviceID:        respData.DeviceID,
		EncryptedAESKey: decodedKey,
		CreatedAt:       respData.CreatedAt,
		UpdatedAt:       respData.UpdatedAt,
	}

	return secretKey, nil
}
