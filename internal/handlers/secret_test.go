package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretSaveHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenDecoder := NewMockSecretTokenDecoder(ctrl)
	mockSecretWriter := NewMockSecretWriter(ctrl)

	handler := NewSecretSaveHTTPHandler(mockTokenDecoder, mockSecretWriter)

	t.Run("success save secret", func(t *testing.T) {
		reqBody := SecretResponse{
			UserID:           "user1",
			SecretName:       "secret1",
			SecretType:       "password",
			EncryptedPayload: []byte("encrypted"),
			Nonce:            []byte("nonce"),
			Meta:             "{}",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/save-secret", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mockTokenDecoder.EXPECT().GetFromRequest(req).Return("token", nil)
		mockTokenDecoder.EXPECT().Parse("token").Return("secretID1", "deviceID1", nil)
		mockSecretWriter.EXPECT().Save(
			gomock.Any(),
			"secretID1",
			"user1",
			"secret1",
			"password",
			[]byte("encrypted"),
			[]byte("nonce"),
			"{}",
		).Return(nil)

		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/save-secret", nil)
		w := httptest.NewRecorder()

		mockTokenDecoder.EXPECT().GetFromRequest(req).Return("", errors.New("no token"))

		handler(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})
}

func TestNewSecretGetHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenDecoder := NewMockSecretTokenDecoder(ctrl)
	mockSecretReader := NewMockSecretReader(ctrl)

	handler := NewSecretGetHTTPHandler(mockTokenDecoder, mockSecretReader)

	t.Run("success get secret", func(t *testing.T) {
		secret := &models.SecretDB{
			SecretID:         "id1",
			UserID:           "user1",
			SecretName:       "secret1",
			SecretType:       "password",
			EncryptedPayload: []byte("enc"),
			Nonce:            []byte("nonce"),
			Meta:             "{}",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}

		req := httptest.NewRequest(http.MethodGet, "/get-secret?secret_name=secret1", nil)
		w := httptest.NewRecorder()

		mockTokenDecoder.EXPECT().GetFromRequest(req).Return("token", nil)
		mockTokenDecoder.EXPECT().Parse("token").Return("secretID", "user1", nil)
		mockSecretReader.EXPECT().Get(gomock.Any(), "user1", "secret1").Return(secret, nil)

		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		var resp SecretResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Equal(t, "id1", resp.SecretID)
		assert.Equal(t, "user1", resp.UserID)
		assert.Equal(t, "secret1", resp.SecretName)
	})
}

func TestNewSecretListHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenDecoder := NewMockSecretTokenDecoder(ctrl)
	mockSecretReader := NewMockSecretReader(ctrl)

	handler := NewSecretListHTTPHandler(mockTokenDecoder, mockSecretReader)

	t.Run("success list secrets", func(t *testing.T) {
		secrets := []*models.SecretDB{
			{
				SecretID:         "id1",
				UserID:           "user1",
				SecretName:       "secret1",
				SecretType:       "password",
				EncryptedPayload: []byte("enc"),
				Nonce:            []byte("nonce"),
				Meta:             "{}",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			{
				SecretID:         "id2",
				UserID:           "user1",
				SecretName:       "secret2",
				SecretType:       "note",
				EncryptedPayload: []byte("enc2"),
				Nonce:            []byte("nonce2"),
				Meta:             "{}",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
		}

		req := httptest.NewRequest(http.MethodGet, "/list-secrets", nil)
		w := httptest.NewRecorder()

		mockTokenDecoder.EXPECT().GetFromRequest(req).Return("token", nil)
		mockTokenDecoder.EXPECT().Parse("token").Return("secretID", "user1", nil)
		mockSecretReader.EXPECT().List(gomock.Any(), "user1").Return(secrets, nil)

		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Result().StatusCode)

		var resp []SecretResponse
		json.NewDecoder(w.Body).Decode(&resp)
		assert.Len(t, resp, 2)
		assert.Equal(t, "secret1", resp[0].SecretName)
		assert.Equal(t, "secret2", resp[1].SecretName)
	})
}
