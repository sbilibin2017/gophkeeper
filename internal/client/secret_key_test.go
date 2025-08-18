package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSecretKeyHTTPClient_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := models.SecretKeyResponse{
			SecretKeyID:     "a1b2c3d4-e5f6-7890-abcd-1234567890ef",
			SecretID:        "secret-12345",
			DeviceID:        "device-67890",
			EncryptedAESKey: "U2FsdGVkX1+abcd1234efgh5678ijkl90==",
			CreatedAt:       time.Date(2025, 8, 17, 12, 34, 56, 0, time.UTC),
			UpdatedAt:       time.Date(2025, 8, 17, 12, 45, 0, 0, time.UTC),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fmt.Sprintf("/get/%s", expected.SecretKeyID), r.URL.Path)
			assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		client := NewSecretKeyHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.Get(context.Background(), "mytoken", expected.SecretKeyID)
		assert.NoError(t, err)
		assert.Equal(t, expected.SecretKeyID, result.SecretKeyID)
		assert.Equal(t, expected.SecretID, result.SecretID)
		assert.Equal(t, expected.DeviceID, result.DeviceID)
		assert.Equal(t, expected.EncryptedAESKey, result.EncryptedAESKey)
		assert.True(t, expected.CreatedAt.Equal(result.CreatedAt))
		assert.True(t, expected.UpdatedAt.Equal(result.UpdatedAt))
	})

	t.Run("invalid json", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("not a json"))
		}))
		defer server.Close()

		client := NewSecretKeyHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.Get(context.Background(), "mytoken", "sk123")
		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("http error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewSecretKeyHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.Get(context.Background(), "mytoken", "sk123")
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http error")
	})
}
