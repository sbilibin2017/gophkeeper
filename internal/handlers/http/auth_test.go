package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		expectedStatus     int
		expectedAuthHeader string
		expectedBody       string
		mockSetup          func(ctrl *gomock.Controller) (Registerer, JWTGenerator)
	}{
		{
			name:               "success",
			requestBody:        RegisterRequest{Username: "alice", Password: "pass123"},
			expectedStatus:     http.StatusOK,
			expectedAuthHeader: "Bearer sometoken",
			mockSetup: func(ctrl *gomock.Controller) (Registerer, JWTGenerator) {
				mockRegisterer := NewMockRegisterer(ctrl)
				mockJWTGen := NewMockJWTGenerator(ctrl)

				mockRegisterer.EXPECT().
					Register(gomock.Any(), "alice", "pass123").
					Return(nil).
					Times(1)

				mockJWTGen.EXPECT().
					Generate("alice").
					Return("sometoken", nil).
					Times(1)

				return mockRegisterer, mockJWTGen
			},
		},
		{
			name:           "invalid json",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body\n",
			mockSetup: func(ctrl *gomock.Controller) (Registerer, JWTGenerator) {
				// No calls expected
				return nil, nil
			},
		},
		{
			name:           "user already exists",
			requestBody:    RegisterRequest{Username: "bob", Password: "pass123"},
			expectedStatus: http.StatusConflict,
			expectedBody:   services.ErrUserAlreadyExists.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (Registerer, JWTGenerator) {
				mockRegisterer := NewMockRegisterer(ctrl)
				mockJWTGen := NewMockJWTGenerator(ctrl)

				mockRegisterer.EXPECT().
					Register(gomock.Any(), "bob", "pass123").
					Return(services.ErrUserAlreadyExists).
					Times(1)

				// JWTGen not expected to be called
				return mockRegisterer, mockJWTGen
			},
		},
		{
			name:           "internal server error on register",
			requestBody:    RegisterRequest{Username: "charlie", Password: "pass123"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error\n",
			mockSetup: func(ctrl *gomock.Controller) (Registerer, JWTGenerator) {
				mockRegisterer := NewMockRegisterer(ctrl)
				mockJWTGen := NewMockJWTGenerator(ctrl)

				mockRegisterer.EXPECT().
					Register(gomock.Any(), "charlie", "pass123").
					Return(errors.New("db error")).
					Times(1)

				return mockRegisterer, mockJWTGen
			},
		},
		{
			name:           "jwt generation fails",
			requestBody:    RegisterRequest{Username: "dave", Password: "pass123"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to generate token\n",
			mockSetup: func(ctrl *gomock.Controller) (Registerer, JWTGenerator) {
				mockRegisterer := NewMockRegisterer(ctrl)
				mockJWTGen := NewMockJWTGenerator(ctrl)

				mockRegisterer.EXPECT().
					Register(gomock.Any(), "dave", "pass123").
					Return(nil).
					Times(1)

				mockJWTGen.EXPECT().
					Generate("dave").
					Return("", errors.New("jwt error")).
					Times(1)

				return mockRegisterer, mockJWTGen
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRegisterer, mockJWTGen := tt.mockSetup(ctrl)
			handler := NewRegisterHandler(mockRegisterer, mockJWTGen)

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedAuthHeader != "" {
				assert.Equal(t, tt.expectedAuthHeader, rec.Header().Get("Authorization"))
			}
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		expectedStatus     int
		expectedAuthHeader string
		expectedBody       string
		mockSetup          func(ctrl *gomock.Controller) (Authenticator, JWTGenerator)
	}{
		{
			name:               "success",
			requestBody:        LoginRequest{Username: "alice", Password: "pass123"},
			expectedStatus:     http.StatusOK,
			expectedAuthHeader: "Bearer sometoken",
			mockSetup: func(ctrl *gomock.Controller) (Authenticator, JWTGenerator) {
				mockAuthenticator := NewMockAuthenticator(ctrl)
				mockJWTGen := NewMockJWTGenerator(ctrl)

				mockAuthenticator.EXPECT().
					Authenticate(gomock.Any(), "alice", "pass123").
					Return(nil).
					Times(1)

				mockJWTGen.EXPECT().
					Generate("alice").
					Return("sometoken", nil).
					Times(1)

				return mockAuthenticator, mockJWTGen
			},
		},
		{
			name:           "invalid json",
			requestBody:    "invalid-json",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body\n",
			mockSetup: func(ctrl *gomock.Controller) (Authenticator, JWTGenerator) {
				return nil, nil
			},
		},
		{
			name:           "invalid username or password",
			requestBody:    LoginRequest{Username: "bob", Password: "wrongpass"},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   services.ErrInvalidData.Error() + "\n",
			mockSetup: func(ctrl *gomock.Controller) (Authenticator, JWTGenerator) {
				mockAuthenticator := NewMockAuthenticator(ctrl)
				mockJWTGen := NewMockJWTGenerator(ctrl)

				mockAuthenticator.EXPECT().
					Authenticate(gomock.Any(), "bob", "wrongpass").
					Return(services.ErrInvalidData).
					Times(1)

				return mockAuthenticator, mockJWTGen
			},
		},
		{
			name:           "internal server error on authenticate",
			requestBody:    LoginRequest{Username: "charlie", Password: "pass123"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error\n",
			mockSetup: func(ctrl *gomock.Controller) (Authenticator, JWTGenerator) {
				mockAuthenticator := NewMockAuthenticator(ctrl)
				mockJWTGen := NewMockJWTGenerator(ctrl)

				mockAuthenticator.EXPECT().
					Authenticate(gomock.Any(), "charlie", "pass123").
					Return(errors.New("db error")).
					Times(1)

				return mockAuthenticator, mockJWTGen
			},
		},
		{
			name:           "jwt generation fails",
			requestBody:    LoginRequest{Username: "dave", Password: "pass123"},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "failed to generate token\n",
			mockSetup: func(ctrl *gomock.Controller) (Authenticator, JWTGenerator) {
				mockAuthenticator := NewMockAuthenticator(ctrl)
				mockJWTGen := NewMockJWTGenerator(ctrl)

				mockAuthenticator.EXPECT().
					Authenticate(gomock.Any(), "dave", "pass123").
					Return(nil).
					Times(1)

				mockJWTGen.EXPECT().
					Generate("dave").
					Return("", errors.New("jwt error")).
					Times(1)

				return mockAuthenticator, mockJWTGen
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAuthenticator, mockJWTGen := tt.mockSetup(ctrl)
			handler := NewLoginHandler(mockAuthenticator, mockJWTGen)

			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedAuthHeader != "" {
				assert.Equal(t, tt.expectedAuthHeader, rec.Header().Get("Authorization"))
			}
			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}
		})
	}
}
