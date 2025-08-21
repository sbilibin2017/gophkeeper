package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/require"
)

func TestNewDeviceGetHTTPHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenDecoder := NewMockTokenDecoder(ctrl)
	mockDeviceGetter := NewMockDeviceGetter(ctrl)

	handler := NewDeviceGetHTTPHandler(mockTokenDecoder, mockDeviceGetter)

	userID := "user123"
	deviceID := "device123"
	claims := &models.Claims{
		UserID:   userID,
		DeviceID: deviceID,
	}

	device := &models.DeviceDB{
		UserID:    userID,
		DeviceID:  deviceID,
		PublicKey: "pubkey123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name               string
		setupMocks         func()
		expectedStatusCode int
		expectedBody       interface{}
	}{
		{
			name: "success",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
				mockDeviceGetter.EXPECT().Get(gomock.Any(), userID, deviceID).Return(device, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody: models.DeviceResponse{
				DeviceID:  deviceID,
				UserID:    userID,
				PublicKey: "pubkey123",
				CreatedAt: device.CreatedAt,
				UpdatedAt: device.UpdatedAt,
			},
		},
		{
			name: "bad request token extraction",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("fail"))
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "unauthorized token parsing",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(nil, errors.New("invalid"))
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "device not found",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
				mockDeviceGetter.EXPECT().Get(gomock.Any(), userID, deviceID).Return(nil, nil)
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "internal error from device getter",
			setupMocks: func() {
				mockTokenDecoder.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockTokenDecoder.EXPECT().Parse("token").Return(claims, nil)
				mockDeviceGetter.EXPECT().Get(gomock.Any(), userID, deviceID).Return(nil, errors.New("db error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()

			req := httptest.NewRequest(http.MethodGet, "/get", nil)
			w := httptest.NewRecorder()

			handler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.expectedStatusCode, res.StatusCode)

			if tt.expectedStatusCode == http.StatusOK {
				var resp models.DeviceResponse
				err := json.NewDecoder(res.Body).Decode(&resp)
				require.NoError(t, err)

				expected := tt.expectedBody.(models.DeviceResponse)
				require.Equal(t, expected.DeviceID, resp.DeviceID)
				require.Equal(t, expected.UserID, resp.UserID)
				require.Equal(t, expected.PublicKey, resp.PublicKey)
				require.True(t, expected.CreatedAt.Equal(resp.CreatedAt))
				require.True(t, expected.UpdatedAt.Equal(resp.UpdatedAt))
			}
		})
	}
}
