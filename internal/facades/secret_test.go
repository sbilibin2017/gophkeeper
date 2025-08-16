package facades

import (
	"context"
	"encoding/base64"
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

func TestSecretHTTPFacade_Save(t *testing.T) {
	userID := "mocked-user"
	secretID := "secret-1"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+userID {
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

	err := facade.Save(context.Background(),
		secretID, userID, "test", "note", []byte("payload"), []byte("nonce"), "meta",
	)
	assert.NoError(t, err)
}

func decodeBase64(t *testing.T, s string) []byte {
	data, err := base64.StdEncoding.DecodeString(s)
	assert.NoError(t, err)
	return data
}

func TestSecretHTTPFacade_Get(t *testing.T) {
	userID := "mocked-user"
	secretName := "test-secret"

	updatedAt, _ := time.Parse(time.RFC3339, "2025-08-16T12:00:00Z")

	secretDB := &models.SecretDB{
		SecretID:         "123",
		UserID:           userID,
		SecretName:       secretName,
		SecretType:       "note",
		EncryptedPayload: base64.StdEncoding.EncodeToString([]byte("payload")),
		Nonce:            base64.StdEncoding.EncodeToString([]byte("nonce")),
		Meta:             "meta",
		UpdatedAt:        updatedAt,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+userID {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/get/"+secretName {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secretDB)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	secretResp, err := facade.Get(context.Background(), userID, secretName)
	assert.NoError(t, err)

	payload := decodeBase64(t, secretResp.EncryptedPayload)
	nonceDecoded := decodeBase64(t, secretResp.Nonce)

	assert.Equal(t, secretName, secretResp.SecretName)
	assert.Equal(t, "note", secretResp.SecretType)
	assert.Equal(t, []byte("payload"), payload)
	assert.Equal(t, []byte("nonce"), nonceDecoded)
	assert.Equal(t, "meta", secretResp.Meta)
}

func TestSecretHTTPFacade_List(t *testing.T) {
	userID := "mocked-user"

	updatedAt, _ := time.Parse(time.RFC3339, "2025-08-16T12:00:00Z")

	secretDBList := []*models.SecretDB{
		{
			SecretID:         "123",
			UserID:           userID,
			SecretName:       "test-secret",
			SecretType:       "note",
			EncryptedPayload: base64.StdEncoding.EncodeToString([]byte("payload")),
			Nonce:            base64.StdEncoding.EncodeToString([]byte("nonce")),
			Meta:             "meta",
			UpdatedAt:        updatedAt,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+userID {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path == "/list" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(secretDBList)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	secretsResp, err := facade.List(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, secretsResp, 1)

	payload := decodeBase64(t, secretsResp[0].EncryptedPayload)
	nonceDecoded := decodeBase64(t, secretsResp[0].Nonce)

	assert.Equal(t, "test-secret", secretsResp[0].SecretName)
	assert.Equal(t, "note", secretsResp[0].SecretType)
	assert.Equal(t, []byte("payload"), payload)
	assert.Equal(t, []byte("nonce"), nonceDecoded)
	assert.Equal(t, "meta", secretsResp[0].Meta)
}

func TestSecretHTTPFacade_Save_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	err := facade.Save(context.Background(),
		"123", "user", "name", "note", []byte{}, []byte{}, "",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http error")
}

func TestSecretHTTPFacade_Get_Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewSecretHTTPFacade(client)

	secret, err := facade.Get(context.Background(), "user", "nonexistent")
	assert.Error(t, err)
	assert.Nil(t, secret)
}

func TestSecretHTTPFacade_List_Error(t *testing.T) {
	client := resty.New().SetBaseURL("http://127.0.0.1:0")
	facade := NewSecretHTTPFacade(client)

	secrets, err := facade.List(context.Background(), "user")
	assert.Error(t, err)
	assert.Nil(t, secrets)
}
