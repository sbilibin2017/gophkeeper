package facades

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestSecretHTTPFacade_Save(t *testing.T) {
	token := "mocked-token"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/save" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{}`)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	err := facade.Save(context.Background(), token, &models.SecretRequest{
		SecretName:       "test",
		SecretType:       "note",
		EncryptedPayload: []byte("payload"),
		Nonce:            []byte("nonce"),
		Meta:             "meta",
	})
	assert.NoError(t, err)
}

func decodeBase64(t *testing.T, s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	assert.NoError(t, err)
	return data
}

func TestSecretHTTPFacade_Get(t *testing.T) {
	token := "mocked-token"
	secretName := "test-secret"
	secretType := "note"
	encryptedPayload := []byte("payload")
	nonce := []byte("nonce")
	meta := "meta"

	respObj := models.SecretResponse{
		SecretID:         "123",
		UserID:           "u1",
		SecretName:       secretName,
		SecretType:       secretType,
		EncryptedPayload: base64.StdEncoding.EncodeToString(encryptedPayload),
		Nonce:            base64.StdEncoding.EncodeToString(nonce),
		Meta:             meta,
		UpdatedAt:        "2025-08-16T12:00:00Z",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/get/"+secretName {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(respObj)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	secretResp, err := facade.Get(context.Background(), token, secretName)
	assert.NoError(t, err)

	payload := decodeBase64(t, secretResp.EncryptedPayload)
	nonceDecoded := decodeBase64(t, secretResp.Nonce)

	assert.Equal(t, secretName, secretResp.SecretName)
	assert.Equal(t, secretType, secretResp.SecretType)
	assert.Equal(t, encryptedPayload, payload)
	assert.Equal(t, nonce, nonceDecoded)
	assert.Equal(t, meta, secretResp.Meta)
}

func TestSecretHTTPFacade_List(t *testing.T) {
	token := "mocked-token"
	secretName := "test-secret"
	secretType := "note"
	encryptedPayload := []byte("payload")
	nonce := []byte("nonce")
	meta := "meta"

	respObj := []models.SecretResponse{
		{
			SecretID:         "123",
			UserID:           "u1",
			SecretName:       secretName,
			SecretType:       secretType,
			EncryptedPayload: base64.StdEncoding.EncodeToString(encryptedPayload),
			Nonce:            base64.StdEncoding.EncodeToString(nonce),
			Meta:             meta,
			UpdatedAt:        "2025-08-16T12:00:00Z",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/list" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(respObj)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	secretsResp, err := facade.List(context.Background(), token)
	assert.NoError(t, err)
	assert.Len(t, secretsResp, 1)

	payload := decodeBase64(t, secretsResp[0].EncryptedPayload)
	nonceDecoded := decodeBase64(t, secretsResp[0].Nonce)

	assert.Equal(t, secretName, secretsResp[0].SecretName)
	assert.Equal(t, secretType, secretsResp[0].SecretType)
	assert.Equal(t, encryptedPayload, payload)
	assert.Equal(t, nonce, nonceDecoded)
	assert.Equal(t, meta, secretsResp[0].Meta)
}

func TestSecretHTTPFacade_Save_Error(t *testing.T) {
	// Сервер возвращает 500
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	err := facade.Save(context.Background(), "token", &models.SecretRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http error")
}

func TestSecretHTTPFacade_Get_Error(t *testing.T) {
	// Сервер возвращает 404
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	secret, err := facade.Get(context.Background(), "token", "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, secret)
}

func TestSecretHTTPFacade_List_Error(t *testing.T) {
	// Сервер недоступен (симулируем ошибку соединения)
	client := resty.New().SetBaseURL("http://127.0.0.1:0") // неверный адрес
	facade := NewSecretHTTPFacade(client)

	secrets, err := facade.List(context.Background(), "token")
	assert.Error(t, err)
	assert.Nil(t, secrets)
}

func TestSecretHTTPFacade_Unauthorized(t *testing.T) {
	token := "wrong-token"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	err := facade.Save(context.Background(), token, &models.SecretRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http error")

	secret, err := facade.Get(context.Background(), token, "any")
	assert.Error(t, err)
	assert.Nil(t, secret)

	secrets, err := facade.List(context.Background(), token)
	assert.Error(t, err)
	assert.Nil(t, secrets)
}

// roundTripperFunc позволяет мокать сетевые ошибки
type roundTripperFunc3 func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc3) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSecretHTTPFacade_Save_RequestError(t *testing.T) {
	client := resty.New()
	client.SetTransport(roundTripperFunc3(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	}))

	facade := NewSecretHTTPFacade(client)

	err := facade.Save(context.Background(), "token", &models.SecretRequest{
		SecretName:       "test",
		SecretType:       "note",
		EncryptedPayload: []byte("payload"),
		Nonce:            []byte("nonce"),
		Meta:             "meta",
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

func TestSecretHTTPFacade_Get_RequestError(t *testing.T) {
	client := resty.New()
	client.SetTransport(roundTripperFunc3(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("network error")
	}))

	facade := NewSecretHTTPFacade(client)

	resp, err := facade.Get(context.Background(), "token", "secretName")
	assert.Nil(t, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}
