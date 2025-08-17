package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewSecretKeyGetHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenDecoder := NewMockSecretKeyTokenDecoder(ctrl)
	mockSecretKeyGetter := NewMockSecretKeyGetter(ctrl)

	handler := NewSecretKeyGetHTTPHandler(mockTokenDecoder, mockSecretKeyGetter)

	const (
		secretID = "secret123"
		deviceID = "device123"
		token    = "validtoken"
	)

	secretKey := &models.SecretKeyDB{
		SecretKeyID:     "key123",
		SecretID:        secretID,
		DeviceID:        deviceID,
		EncryptedAESKey: []byte("encryptedkey"),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	tests := []struct {
		name               string
		setupMocks         func()
		expectedStatusCode int
	}{
		{
			name: "success",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return(token, nil)
				mockTokenDecoder.EXPECT().Parse(token).Return(secretID, deviceID, nil)
				mockSecretKeyGetter.EXPECT().Get(gomock.Any(), secretID, deviceID).Return(secretKey, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "bad_request_token_missing",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("no token"))
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "unauthorized_invalid_token",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return(token, nil)
				mockTokenDecoder.EXPECT().Parse(token).Return("", "", errors.New("invalid token"))
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "not_found_secret_key_missing",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return(token, nil)
				mockTokenDecoder.EXPECT().Parse(token).Return(secretID, deviceID, nil)
				mockSecretKeyGetter.EXPECT().Get(gomock.Any(), secretID, deviceID).Return(nil, nil)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "internal_error",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return(token, nil)
				mockTokenDecoder.EXPECT().Parse(token).Return(secretID, deviceID, nil)
				mockSecretKeyGetter.EXPECT().Get(gomock.Any(), secretID, deviceID).Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/get-secret-key", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatusCode, rec.Result().StatusCode)
		})
	}
}
