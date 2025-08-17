package facades

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestSecretKeyHTTPFacade_Get(t *testing.T) {
	userToken := "mocked-token"
	secretKeyID := "key123"
	secretID := "secret123"
	deviceID := "device123"

	// Base64 закодированный ключ
	rawKey := []byte("my-secret-aes-key")
	encryptedAESKey := base64.StdEncoding.EncodeToString(rawKey)

	// Мок HTTP-сервера
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+userToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/get" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"secret_key_id":"%s",
				"secret_id":"%s",
				"device_id":"%s",
				"encrypted_aes_key":"%s",
				"created_at":"2025-08-16T12:00:00Z",
				"updated_at":"2025-08-16T12:00:00Z"
			}`, secretKeyID, secretID, deviceID, encryptedAESKey)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretKeyHTTPFacade(client)

	// === Успешный запрос ===
	resp, err := facade.Get(context.Background(), userToken, deviceID)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, secretKeyID, resp.SecretKeyID)
	assert.Equal(t, secretID, resp.SecretID)
	assert.Equal(t, deviceID, resp.DeviceID)
	assert.Equal(t, rawKey, resp.EncryptedAESKey)

	expectedTime, _ := time.Parse(time.RFC3339, "2025-08-16T12:00:00Z")
	assert.Equal(t, expectedTime, resp.CreatedAt)
	assert.Equal(t, expectedTime, resp.UpdatedAt)

	// === Ошибка авторизации ===
	resp, err = facade.Get(context.Background(), "wrong-token", deviceID)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http error: 401 Unauthorized")
}

func TestSecretKeyHTTPFacade_Get_RequestError(t *testing.T) {
	secretKeyID := "key123"
	deviceID := "device123"

	client := resty.New()
	client.SetTransport(roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("network error")
	}))

	facade := NewSecretKeyHTTPFacade(client)

	resp, err := facade.Get(context.Background(), secretKeyID, deviceID)
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

// roundTripperFunc помогает мокать сетевые ошибки
type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
