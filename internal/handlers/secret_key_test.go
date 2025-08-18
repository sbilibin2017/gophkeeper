package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/require"
)

func TestNewSecretKeySaveHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenDecoder := NewMockSecretKeyTokenDecoder(ctrl)
	mockSecretKeyWriter := NewMockSecretKeyWriter(ctrl)

	handler := NewSecretKeySaveHTTPHandler(mockTokenDecoder, mockSecretKeyWriter)

	userID := "user123"
	deviceID := "device123"
	claims := &models.Claims{
		UserID:   userID,
		DeviceID: deviceID,
	}

	reqBody := models.SecretKeyRequest{
		SecretID:        "secret123",
		EncryptedAESKey: "encryptedKey",
	}

	tests := []struct {
		name               string
		setupMocks         func()
		requestBody        interface{}
		expectedStatusCode int
	}{
		{
			name: "success",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
				mockSecretKeyWriter.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
			},
			requestBody:        reqBody,
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "bad request token extraction",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("fail"))
			},
			requestBody:        reqBody,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "unauthorized token parsing",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(nil, errors.New("invalid"))
			},
			requestBody:        reqBody,
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "bad request invalid JSON",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
			},
			requestBody:        "invalid-json",
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "internal error on save",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
				mockSecretKeyWriter.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errors.New("db error"))
			},
			requestBody:        reqBody,
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/save", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			handler(w, req)

			require.Equal(t, tt.expectedStatusCode, w.Result().StatusCode)
		})
	}
}

func TestNewSecretKeyGetHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenDecoder := NewMockSecretKeyTokenDecoder(ctrl)
	mockSecretKeyGetter := NewMockSecretKeyGetter(ctrl)

	handler := NewSecretKeyGetHTTPHandler(mockTokenDecoder, mockSecretKeyGetter)

	userID := "user123"
	deviceID := "device123"
	claims := &models.Claims{
		UserID:   userID,
		DeviceID: deviceID,
	}

	now := time.Now().Truncate(time.Microsecond) // truncate to microseconds
	secretKey := &models.SecretKeyDB{
		SecretKeyID:     "key123",
		SecretID:        "secret123",
		DeviceID:        deviceID,
		EncryptedAESKey: "encryptedKey",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	tests := []struct {
		name               string
		secretID           string
		setupMocks         func()
		expectedStatusCode int
		expectedBody       *models.SecretKeyResponse
	}{
		{
			name:     "success",
			secretID: "secret123",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
				mockSecretKeyGetter.EXPECT().Get(gomock.Any(), "secret123", deviceID).Return(secretKey, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody: &models.SecretKeyResponse{
				SecretKeyID:     secretKey.SecretKeyID,
				SecretID:        secretKey.SecretID,
				DeviceID:        secretKey.DeviceID,
				EncryptedAESKey: secretKey.EncryptedAESKey,
				CreatedAt:       secretKey.CreatedAt,
				UpdatedAt:       secretKey.UpdatedAt,
			},
		},
		{
			name:     "bad request token extraction",
			secretID: "secret123",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("fail"))
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:     "unauthorized token parsing",
			secretID: "secret123",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(nil, errors.New("invalid"))
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name:     "secret not found",
			secretID: "secret123",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
				mockSecretKeyGetter.EXPECT().Get(gomock.Any(), "secret123", deviceID).Return(nil, nil)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name:     "internal error on get",
			secretID: "secret123",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
				mockSecretKeyGetter.EXPECT().Get(gomock.Any(), "secret123", deviceID).Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/get/"+tt.secretID, nil)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext()))
			chi.RouteContext(req.Context()).URLParams.Add("secret-id", tt.secretID)

			w := httptest.NewRecorder()
			handler(w, req)

			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.expectedStatusCode, res.StatusCode)

			if tt.expectedStatusCode == http.StatusOK {
				var resp models.SecretKeyResponse
				err := json.NewDecoder(res.Body).Decode(&resp)
				require.NoError(t, err)

				// Compare field by field, ignoring monotonic clock in time
				require.Equal(t, tt.expectedBody.SecretKeyID, resp.SecretKeyID)
				require.Equal(t, tt.expectedBody.SecretID, resp.SecretID)
				require.Equal(t, tt.expectedBody.DeviceID, resp.DeviceID)
				require.Equal(t, tt.expectedBody.EncryptedAESKey, resp.EncryptedAESKey)
				require.True(t, tt.expectedBody.CreatedAt.Equal(resp.CreatedAt))
				require.True(t, tt.expectedBody.UpdatedAt.Equal(resp.UpdatedAt))
			}
		})
	}
}
