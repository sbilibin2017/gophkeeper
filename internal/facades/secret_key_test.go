package facades

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSecretKeyHTTPFacade_Get(t *testing.T) {
	token := "mocked-token"
	secretKeyID := "key123"
	secretID := "secret123"
	deviceID := "device123"
	encryptedAESKey := "encrypted-key"

	// Тестовый HTTP сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		expectedPath := fmt.Sprintf("/get/%s", secretKeyID)
		if r.URL.Path == expectedPath {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"secret_key_id":"%s",
				"secret_id":"%s",
				"device_id":"%s",
				"encrypted_aes_key":"%s",
				"updated_at":"2025-08-16T12:00:00Z"
			}`, secretKeyID, secretID, deviceID, encryptedAESKey)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretKeyHTTPFacade(client)

	// Проверка успешного запроса
	resp, err := facade.Get(context.Background(), token, secretKeyID)
	assert.NoError(t, err)
	assert.Equal(t, secretKeyID, resp.SecretKeyID)
	assert.Equal(t, secretID, resp.SecretID)
	assert.Equal(t, deviceID, resp.DeviceID)
	assert.Equal(t, encryptedAESKey, resp.EncryptedAESKey)
	assert.Equal(t, time.Date(2025, 8, 16, 12, 0, 0, 0, time.UTC), resp.UpdatedAt)

	// Проверка ошибки авторизации
	_, err = facade.Get(context.Background(), "wrong-token", secretKeyID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http error: 401 Unauthorized")
}

func TestSecretKeyHTTPFacade_Get_RequestError(t *testing.T) {
	token := "mocked-token"
	secretID := "key123"

	// Создаём клиент resty с кастомным транспортом, который всегда возвращает ошибку
	client := resty.New()
	client.SetTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	}))

	facade := NewSecretKeyHTTPFacade(client)

	resp, err := facade.Get(context.Background(), token, secretID)

	assert.Nil(t, resp)
	assert.Error(t, err)

}

// Вспомогательная обёртка для реализации http.RoundTripper
type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
