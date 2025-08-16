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
	userID := "user123"
	deviceID := "device123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+userID { // используем userID, т.к. SetAuthToken(userID)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/get" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"device_id":"%s",
				"user_id":"%s",
				"public_key":"pubkey",
				"created_at":"2025-08-16T12:00:00Z",
				"updated_at":"2025-08-16T12:00:00Z"
			}`, deviceID, userID)
			return
		}

		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewDeviceHTTPFacade(client)

	// Успешный запрос
	resp, err := facade.Get(context.Background(), userID, deviceID)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, deviceID, resp.DeviceID)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, "pubkey", resp.PublicKey)

	// Ошибка авторизации
	resp, err = facade.Get(context.Background(), "wrong-user", deviceID)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "http error: 401 Unauthorized")
}

func TestDeviceHTTPFacade_Get_RequestError(t *testing.T) {
	userID := "user123"
	deviceID := "device123"

	client := resty.New()
	client.SetTransport(roundTripperFunc2(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	}))

	facade := NewDeviceHTTPFacade(client)

	resp, err := facade.Get(context.Background(), userID, deviceID)
	assert.Nil(t, resp)
	assert.Error(t, err)
}

// Вспомогательная обёртка для реализации http.RoundTripper
type roundTripperFunc2 func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc2) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
