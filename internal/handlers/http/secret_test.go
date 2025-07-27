package http

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestSecretAddHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretAddHandler(mockWriter, mockParser)

	validToken := "valid-token"
	username := "user1"

	mockParser.EXPECT().Parse(validToken).Return(username, nil).Times(1)

	reqBody := map[string]interface{}{
		"secret_name": "name1",
		"secret_type": "type1",
		"ciphertext":  []byte("ciphertext-data"),
		"aes_key_enc": []byte("aeskeyenc-data"),
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	mockWriter.EXPECT().Save(
		gomock.Any(),
		username,
		"name1",
		"type1",
		reqBody["ciphertext"].([]byte),
		reqBody["aes_key_enc"].([]byte),
	).Return(nil).Times(1)

	req := httptest.NewRequest(http.MethodPost, "/secret", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+validToken)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSecretAddHandler_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockSecretWriter(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretAddHandler(mockWriter, mockParser)

	reqBody := `{"secret_name":"name1","secret_type":"type1","ciphertext":"Y2lwaGVydGV4dA==","aes_key_enc":"YWVzS2V5"}`

	req := httptest.NewRequest(http.MethodPost, "/secret", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer invalid-token")

	mockParser.EXPECT().Parse("invalid-token").Return("", errors.New("invalid token")).Times(1)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSecretGetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretGetHandler(mockReader, mockParser)

	validToken := "valid-token"
	username := "user1"
	secretType := "type1"
	secretName := "name1"

	mockParser.EXPECT().Parse(validToken).Return(username, nil).Times(1)

	expectedSecret := &models.Secret{
		SecretOwner: username,
		SecretType:  secretType,
		SecretName:  secretName,
		Ciphertext:  []byte("ciphertext-data"),
		AESKeyEnc:   []byte("aeskeyenc-data"),
	}
	mockReader.EXPECT().Get(gomock.Any(), username, secretType, secretName).Return(expectedSecret, nil).Times(1)

	req := httptest.NewRequest(http.MethodGet, "/secrets/"+secretType+"/"+secretName, nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("secret_type", secretType)
	rctx.URLParams.Add("secret_name", secretName)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var got models.Secret
	err := json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)

	assert.Equal(t, expectedSecret.SecretOwner, got.SecretOwner)
	assert.Equal(t, expectedSecret.SecretType, got.SecretType)
	assert.Equal(t, expectedSecret.SecretName, got.SecretName)
	assert.Equal(t, expectedSecret.Ciphertext, got.Ciphertext)
	assert.Equal(t, expectedSecret.AESKeyEnc, got.AESKeyEnc)
}

func TestSecretGetHandler_MissingParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretGetHandler(mockReader, mockParser)

	validToken := "valid-token"
	username := "user1"

	mockParser.EXPECT().Parse(validToken).Return(username, nil).Times(1)

	req := httptest.NewRequest(http.MethodGet, "/secrets//", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	// no URL params

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSecretListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretListHandler(mockReader, mockParser)

	validToken := "valid-token"
	username := "user1"

	mockParser.EXPECT().Parse(validToken).Return(username, nil).Times(1)

	secrets := []*models.Secret{
		{
			SecretOwner: username,
			SecretName:  "name1",
			SecretType:  "type1",
			Ciphertext:  []byte("ciphertext1"),
			AESKeyEnc:   []byte("aeskeyenc1"),
		},
		{
			SecretOwner: username,
			SecretName:  "name2",
			SecretType:  "type2",
			Ciphertext:  []byte("ciphertext2"),
			AESKeyEnc:   []byte("aeskeyenc2"),
		},
	}

	mockReader.EXPECT().List(gomock.Any(), username).Return(secrets, nil).Times(1)

	req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	var got []*models.Secret
	err := json.NewDecoder(resp.Body).Decode(&got)
	require.NoError(t, err)

	assert.Len(t, got, 2)
	assert.Equal(t, secrets[0].SecretName, got[0].SecretName)
	assert.Equal(t, secrets[1].SecretName, got[1].SecretName)
}

func TestSecretListHandler_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockSecretReader(ctrl)
	mockParser := NewMockJWTParser(ctrl)

	handler := NewSecretListHandler(mockReader, mockParser)

	req := httptest.NewRequest(http.MethodGet, "/secrets", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	mockParser.EXPECT().Parse("invalid-token").Return("", errors.New("invalid token")).Times(1)

	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
