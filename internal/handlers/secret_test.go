package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

// helper to generate token claims
func testClaims() *models.Claims {
	return &models.Claims{
		UserID: "user123",
	}
}

// ------------------------- Test Save Handler -------------------------

func TestNewSecretSaveHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDecoder := NewMockSecretTokenDecoder(ctrl)
	mockWriter := NewMockSecretWriter(ctrl)

	handler := NewSecretSaveHTTPHandler(mockDecoder, mockWriter)

	t.Run("success", func(t *testing.T) {
		metaJSON, _ := json.Marshal(map[string]string{"meta": "value"})

		reqBody := models.SecretRequest{
			SecretName:       "test-secret",
			SecretType:       "password",
			EncryptedPayload: base64.StdEncoding.EncodeToString([]byte("encrypted")),
			Nonce:            base64.StdEncoding.EncodeToString([]byte("nonce")),
			Meta:             string(metaJSON), // assign JSON string
		}
		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()

		mockDecoder.EXPECT().GetFromRequest(req).Return("token", nil)
		mockDecoder.EXPECT().Parse("token").Return(testClaims(), nil)
		mockWriter.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte("bad json")))
		w := httptest.NewRecorder()

		mockDecoder.EXPECT().GetFromRequest(req).Return("token", nil)
		mockDecoder.EXPECT().Parse("token").Return(testClaims(), nil)

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})

	t.Run("bad token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		mockDecoder.EXPECT().GetFromRequest(req).Return("", errors.New("no token"))

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
	})
}

// ------------------------- Test Get Handler -------------------------

func TestNewSecretGetHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDecoder := NewMockSecretTokenDecoder(ctrl)
	mockReader := NewMockSecretReader(ctrl)

	handler := NewSecretGetHTTPHandler(mockDecoder, mockReader)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/get/secret123", nil)

		// Create a chi RouteContext and add URL param
		rc := chi.NewRouteContext()
		rc.URLParams.Add("secret-id", "secret123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

		w := httptest.NewRecorder()

		mockDecoder.EXPECT().GetFromRequest(req).Return("token", nil)
		mockDecoder.EXPECT().Parse("token").Return(testClaims(), nil)
		metaJSON, _ := json.Marshal(map[string]string{"a": "b"})
		mockReader.EXPECT().Get(gomock.Any(), "user123", "secret123").Return(&models.SecretDB{
			SecretID:         "secret123",
			UserID:           "user123",
			SecretName:       "name",
			SecretType:       "type",
			EncryptedPayload: "encrypted",
			Nonce:            "nonce",
			Meta:             string(metaJSON),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}, nil)

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
}

// ------------------------- Test List Handler -------------------------

func TestNewSecretListHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDecoder := NewMockSecretTokenDecoder(ctrl)
	mockReader := NewMockSecretReader(ctrl)

	handler := NewSecretListHTTPHandler(mockDecoder, mockReader)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/list", nil)
		w := httptest.NewRecorder()

		mockDecoder.EXPECT().GetFromRequest(req).Return("token", nil)
		mockDecoder.EXPECT().Parse("token").Return(testClaims(), nil)
		mockReader.EXPECT().List(gomock.Any(), "user123").Return([]*models.SecretDB{
			{
				SecretID:         "s1",
				UserID:           "user123",
				SecretName:       "n1",
				SecretType:       "t1",
				EncryptedPayload: "e1",
				Nonce:            "n1",
				Meta:             "",
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
		}, nil)

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Result().StatusCode)
	})
}
