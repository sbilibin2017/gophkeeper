package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDeviceHTTPClient_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := models.DeviceResponse{
			DeviceID:  "f47ac10b-58cc-4372-a567-0e02b2c3d479",
			UserID:    "c56a4180-65aa-42ec-a945-5fd21dec0538",
			PublicKey: "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAu7pM4h2...\n-----END PUBLIC KEY-----",
			CreatedAt: time.Date(2025, 8, 17, 12, 34, 56, 0, time.UTC),
			UpdatedAt: time.Date(2025, 8, 17, 12, 45, 0, 0, time.UTC),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/get", r.URL.Path)
			assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json") // <<< fix
			json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		client := NewDeviceHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.Get(context.Background(), "mytoken")
		assert.NoError(t, err)
		assert.Equal(t, expected.DeviceID, result.DeviceID)
		assert.Equal(t, expected.UserID, result.UserID)
		assert.Equal(t, expected.PublicKey, result.PublicKey)
		assert.True(t, expected.CreatedAt.Equal(result.CreatedAt))
		assert.True(t, expected.UpdatedAt.Equal(result.UpdatedAt))
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json") // <<< fix
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not a json"))
		}))
		defer server.Close()

		client := NewDeviceHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.Get(context.Background(), "mytoken")
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
