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

func TestSecretHTTPClient_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		req := &models.SecretRequest{
			UserID:           "user789",
			SecretName:       "my-password",
			SecretType:       "password",
			EncryptedPayload: "SGVsbG8gV29ybGQh",
			Nonce:            "MTIzNDU2Nzg5MA==",
			Meta:             `{"url":"https://example.com"}`,
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/save", r.URL.Path)
			assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
			var body models.SecretRequest
			err := json.NewDecoder(r.Body).Decode(&body)
			assert.NoError(t, err)
			assert.Equal(t, req, &body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewSecretHTTPClient(resty.New().SetBaseURL(server.URL))
		err := client.Save(context.Background(), "mytoken", req)
		assert.NoError(t, err)
	})

	t.Run("http error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewSecretHTTPClient(resty.New().SetBaseURL(server.URL))
		err := client.Save(context.Background(), "mytoken", &models.SecretRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http error")
	})
}

func TestSecretHTTPClient_Get(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := models.SecretResponse{
			SecretID:         "abc123",
			UserID:           "user789",
			SecretName:       "MyBankPassword",
			SecretType:       "password",
			EncryptedPayload: "U2FsdGVkX1+abc123xyz==",
			Nonce:            "bXlOb25jZQ==",
			Meta:             `{"url":"https://example.com"}`,
			CreatedAt:        time.Date(2025, 8, 17, 12, 0, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2025, 8, 17, 12, 30, 0, 0, time.UTC),
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, fmt.Sprintf("/get-secret/%s", expected.SecretID), r.URL.Path)
			assert.Equal(t, "Bearer mytoken", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		client := NewSecretHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.Get(context.Background(), "mytoken", expected.SecretID)
		assert.NoError(t, err)
		assert.Equal(t, expected.SecretID, result.SecretID)
		assert.Equal(t, expected.UserID, result.UserID)
		assert.Equal(t, expected.SecretName, result.SecretName)
		assert.Equal(t, expected.SecretType, result.SecretType)
		assert.Equal(t, expected.EncryptedPayload, result.EncryptedPayload)
		assert.Equal(t, expected.Nonce, result.Nonce)
		assert.Equal(t, expected.Meta, result.Meta)
		assert.True(t, expected.CreatedAt.Equal(result.CreatedAt))
		assert.True(t, expected.UpdatedAt.Equal(result.UpdatedAt))
	})

	t.Run("http error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewSecretHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.Get(context.Background(), "mytoken", "sk123")
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http error")
	})
}

func TestSecretHTTPClient_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		expected := []*models.SecretResponse{
			{
				SecretID:         "abc123",
				UserID:           "user789",
				SecretName:       "MyBankPassword",
				SecretType:       "password",
				EncryptedPayload: "U2FsdGVkX1+abc123xyz==",
				Nonce:            "bXlOb25jZQ==",
				Meta:             `{"url":"https://example.com"}`,
				CreatedAt:        time.Date(2025, 8, 17, 12, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 8, 17, 12, 30, 0, 0, time.UTC),
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/list", r.URL.Path)
			assert.Equal(t, "Bearer user789", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(expected)
		}))
		defer server.Close()

		client := NewSecretHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.List(context.Background(), "user789")
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("http error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := NewSecretHTTPClient(resty.New().SetBaseURL(server.URL))
		result, err := client.List(context.Background(), "user789")
		assert.Nil(t, result)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http error")
	})
}
