package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretAddHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretAddHandler(mockWriter, mockParser)

	// Define a request body value for tests
	secretReq := struct {
		SecretName string `json:"secret_name"`
		SecretType string `json:"secret_type"`
		Ciphertext []byte `json:"ciphertext"`
		AESKeyEnc  []byte `json:"aes_key_enc"`
	}{
		SecretName: "mysecret",
		SecretType: "password",
		Ciphertext: []byte("encrypted"),
		AESKeyEnc:  []byte("key"),
	}

	body, err := json.Marshal(secretReq)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer validtoken")
		w := httptest.NewRecorder()

		mockParser.EXPECT().Parse("validtoken").Return("testuser", nil)
		mockWriter.EXPECT().Save(gomock.Any(), "testuser", secretReq.SecretName, secretReq.SecretType, secretReq.Ciphertext, secretReq.AESKeyEnc).Return(nil)

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("missing auth header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid auth header format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(body))
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("jwt parse error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer invalidtoken")
		w := httptest.NewRecorder()

		mockParser.EXPECT().Parse("invalidtoken").Return("", errors.New("invalid token"))

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid json body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/secrets", strings.NewReader("notjson"))
		req.Header.Set("Authorization", "Bearer validtoken")
		w := httptest.NewRecorder()

		mockParser.EXPECT().Parse("validtoken").Return("testuser", nil)

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("save error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer validtoken")
		w := httptest.NewRecorder()

		mockParser.EXPECT().Parse("validtoken").Return("testuser", nil)
		mockWriter.EXPECT().Save(gomock.Any(), "testuser", secretReq.SecretName, secretReq.SecretType, secretReq.Ciphertext, secretReq.AESKeyEnc).
			Return(errors.New("save failure"))

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestSecretGetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretGetHandler(mockReader, mockParser)

	secret := &models.Secret{
		SecretName: "mysecret",
		SecretType: "password",
		Ciphertext: []byte("encrypteddata"),
		AESKeyEnc:  []byte("keydata"),
	}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/secrets/password/mysecret", nil)
		req = req.WithContext(context.Background())
		req.Header.Set("Authorization", "Bearer validtoken")
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("secret_type", "password")
		rctx.URLParams.Add("secret_name", "mysecret")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockParser.EXPECT().Parse("validtoken").Return("testuser", nil)
		mockReader.EXPECT().Get(gomock.Any(), "testuser", "password", "mysecret").Return(secret, nil)

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var gotSecret models.Secret
		err := json.NewDecoder(resp.Body).Decode(&gotSecret)
		require.NoError(t, err)
		assert.Equal(t, secret.SecretName, gotSecret.SecretName)
		assert.Equal(t, secret.SecretType, gotSecret.SecretType)
	})

	t.Run("missing auth header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/secrets/password/mysecret", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("jwt parse error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/secrets/password/mysecret", nil)
		req.Header.Set("Authorization", "Bearer invalidtoken")
		w := httptest.NewRecorder()

		mockParser.EXPECT().Parse("invalidtoken").Return("", errors.New("invalid token"))

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing URL params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/secrets//", nil)
		req.Header.Set("Authorization", "Bearer validtoken")
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockParser.EXPECT().Parse("validtoken").Return("testuser", nil)

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("get secret error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/secrets/password/mysecret", nil)
		req.Header.Set("Authorization", "Bearer validtoken")
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("secret_type", "password")
		rctx.URLParams.Add("secret_name", "mysecret")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockParser.EXPECT().Parse("validtoken").Return("testuser", nil)
		mockReader.EXPECT().Get(gomock.Any(), "testuser", "password", "mysecret").Return(nil, errors.New("some error"))

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestSecretListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretListHandler(mockReader, mockParser)

	secrets := []*models.Secret{
		{
			SecretName: "secret1",
			SecretType: "password",
		},
		{
			SecretName: "secret2",
			SecretType: "card",
		},
	}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
		req.Header.Set("Authorization", "Bearer validtoken")
		w := httptest.NewRecorder()

		mockParser.EXPECT().Parse("validtoken").Return("testuser", nil)
		mockReader.EXPECT().List(gomock.Any(), "testuser").Return(secrets, nil)

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var gotSecrets []*models.Secret
		err := json.NewDecoder(resp.Body).Decode(&gotSecrets)
		require.NoError(t, err)
		assert.Len(t, gotSecrets, 2)
	})

	t.Run("missing auth header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

}
