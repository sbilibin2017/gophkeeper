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

func TestNewDeviceGetHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenDecoder := NewMockTokenDecoder(ctrl)
	mockDeviceGetter := NewMockDeviceGetter(ctrl)

	handler := NewDeviceGetHTTPHandler(mockTokenDecoder, mockDeviceGetter)

	const (
		userID   = "user123"
		deviceID = "device123"
		token    = "validtoken"
	)

	device := &models.DeviceDB{
		DeviceID:  deviceID,
		UserID:    userID,
		PublicKey: "pubkey",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
				mockTokenDecoder.EXPECT().Parse(token).Return(userID, deviceID, nil)
				mockDeviceGetter.EXPECT().Get(gomock.Any(), userID, deviceID).Return(device, nil)
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
			name: "not_found_device_missing",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return(token, nil)
				mockTokenDecoder.EXPECT().Parse(token).Return(userID, deviceID, nil)
				mockDeviceGetter.EXPECT().Get(gomock.Any(), userID, deviceID).Return(nil, nil)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "internal_error",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return(token, nil)
				mockTokenDecoder.EXPECT().Parse(token).Return(userID, deviceID, nil)
				mockDeviceGetter.EXPECT().Get(gomock.Any(), userID, deviceID).Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/get-device", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatusCode, rec.Result().StatusCode)
		})
	}
}
