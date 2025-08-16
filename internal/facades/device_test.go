package facades

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestDeviceHTTPFacade_Get(t *testing.T) {
	token := "mocked-token"
	deviceID := "device123"

	// Тестовый HTTP сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/get" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{
    "device_id":"device123",
    "user_id":"user123",
    "public_key":"pubkey",
    "created_at":"2025-08-16T12:00:00Z",
    "updated_at":"2025-08-16T12:00:00Z"
}`)
			return
		}

		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewDeviceHTTPFacade(client)

	// Успешный запрос
	resp, err := facade.Get(context.Background(), token)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, deviceID, resp.DeviceID)
	assert.Equal(t, "user123", resp.UserID)
	assert.Equal(t, "pubkey", resp.PublicKey)

	// Ошибка авторизации
	resp, err = facade.Get(context.Background(), "wrong-token")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "http error: 401 Unauthorized")
}

func TestDeviceHTTPFacade_Get_RequestError(t *testing.T) {
	token := "mocked-token"

	// Создаём клиент resty с кастомным транспортом, который всегда возвращает ошибку
	client := resty.New()
	client.SetTransport(roundTripperFunc2(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	}))

	facade := NewDeviceHTTPFacade(client)

	resp, err := facade.Get(context.Background(), token)

	assert.Nil(t, resp)
	assert.Error(t, err)
}

// Вспомогательная обёртка для реализации http.RoundTripper
type roundTripperFunc2 func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc2) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
