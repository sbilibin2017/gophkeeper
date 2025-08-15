package facades

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GetTokenFromRestyResponse извлекает токен из заголовка Authorization
func GetTokenFromRestyResponse(resp *resty.Response) (string, error) {
	authHeader := resp.Header().Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	const prefix = "Bearer "
	if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	token := authHeader[len(prefix):]
	if token == "" {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	return token, nil
}

func TestAuthHTTPFacade_Register_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/register" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Authorization", "Bearer test-token")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock-priv-key"))
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client, GetTokenFromRestyResponse)

	privKey, token, err := facade.Register(context.Background(), "user1", "pass1", "device1")
	require.NoError(t, err)
	assert.Equal(t, "mock-priv-key", string(privKey))
	assert.Equal(t, "test-token", token)
}

func TestAuthHTTPFacade_Register_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client, GetTokenFromRestyResponse)

	privKey, token, err := facade.Register(context.Background(), "user1", "pass1", "device1")
	require.Error(t, err)
	assert.Nil(t, privKey)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "registration failed with status 500")
}

func TestAuthHTTPFacade_Register_NoAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock-priv-key"))
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client, GetTokenFromRestyResponse)

	privKey, token, err := facade.Register(context.Background(), "user1", "pass1", "device1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing Authorization")
	assert.Nil(t, privKey)
	assert.Empty(t, token)
}

func TestAuthHTTPFacade_Register_InvalidAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "InvalidTokenFormat")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock-priv-key"))
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client, GetTokenFromRestyResponse)

	privKey, token, err := facade.Register(context.Background(), "user1", "pass1", "device1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Authorization header format")
	assert.Nil(t, privKey)
	assert.Empty(t, token)
}

func TestAuthHTTPFacade_Register_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer ")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mock-priv-key"))
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	facade := NewAuthHTTPFacade(client, GetTokenFromRestyResponse)

	privKey, token, err := facade.Register(context.Background(), "user1", "pass1", "device1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Authorization header format")
	assert.Nil(t, privKey)
	assert.Empty(t, token)
}

func TestAuthHTTPFacade_Register_RequestError(t *testing.T) {
	// Создаем клиент Resty с невалидным BaseURL, чтобы вызов Post сразу вернул ошибку
	client := resty.New().SetBaseURL("http://invalid-host")

	facade := NewAuthHTTPFacade(client, GetTokenFromRestyResponse)

	privKey, token, err := facade.Register(context.Background(), "user1", "pass1", "device1")
	require.Error(t, err)

	assert.Nil(t, privKey)
	assert.Empty(t, token)
}
