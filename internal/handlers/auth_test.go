package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTP Tests

func TestHTTPHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)
	usernameValidator := func(u string) error {
		if u == "baduser" {
			return errors.New("bad username")
		}
		return nil
	}
	passwordValidator := func(p string) error {
		if p == "badpass" {
			return errors.New("bad password")
		}
		return nil
	}

	handler := NewHTTPHandler(mockSvc, usernameValidator, passwordValidator)

	tests := []struct {
		name               string
		body               string
		mockReturnToken    string
		mockReturnError    error
		expectedStatusCode int
		expectedHeader     string
	}{
		{"Successful registration", `{"username":"testuser","password":"testpass"}`, "token123", nil, http.StatusOK, "Bearer token123"},
		{"Invalid JSON body", `{"username":"testuser","password":}`, "", nil, http.StatusBadRequest, ""},
		{"Invalid username", `{"username":"baduser","password":"testpass"}`, "", nil, http.StatusBadRequest, ""},
		{"Invalid password", `{"username":"testuser","password":"badpass"}`, "", nil, http.StatusBadRequest, ""},
		{"User already exists", `{"username":"testuser","password":"testpass"}`, "", services.ErrUserAlreadyExists, http.StatusConflict, ""},
		{"Internal server error", `{"username":"testuser","password":"testpass"}`, "", errors.New("error"), http.StatusInternalServerError, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockReturnError != nil || tt.mockReturnToken != "" {
				mockSvc.EXPECT().
					Register(gomock.Any(), gomock.AssignableToTypeOf(&models.User{})).
					Return(tt.mockReturnToken, tt.mockReturnError).
					Times(1)
			} else {
				mockSvc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Times(0)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, resp.Header.Get("Authorization"))
			}
		})
	}
}

func TestHTTPHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)
	handler := NewHTTPHandler(mockSvc, nil, nil) // No validators for login

	tests := []struct {
		name               string
		body               string
		mockReturnToken    string
		mockReturnError    error
		expectedStatusCode int
		expectedHeader     string
	}{
		{"Successful login", `{"username":"testuser","password":"testpass"}`, "token123", nil, http.StatusOK, "Bearer token123"},
		{"Invalid JSON body", `{"username":"testuser","password":}`, "", nil, http.StatusBadRequest, ""},
		{"Invalid credentials", `{"username":"testuser","password":"wrongpass"}`, "", services.ErrInvalidData, http.StatusUnauthorized, ""},
		{"Internal server error", `{"username":"testuser","password":"testpass"}`, "", errors.New("error"), http.StatusInternalServerError, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockReturnError != nil || tt.mockReturnToken != "" {
				mockSvc.EXPECT().
					Authenticate(gomock.Any(), gomock.AssignableToTypeOf(&models.User{})).
					Return(tt.mockReturnToken, tt.mockReturnError).
					Times(1)
			} else {
				mockSvc.EXPECT().
					Authenticate(gomock.Any(), gomock.Any()).
					Times(0)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			resp := w.Result()
			assert.Equal(t, tt.expectedStatusCode, resp.StatusCode)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, resp.Header.Get("Authorization"))
			}
		})
	}
}

// gRPC Tests

func TestGRPCHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)

	usernameValidator := func(u string) error {
		if u == "baduser" {
			return errors.New("invalid username")
		}
		return nil
	}

	passwordValidator := func(p string) error {
		if p == "badpass" {
			return errors.New("invalid password")
		}
		return nil
	}

	handler := NewGRPCHandler(mockSvc, usernameValidator, passwordValidator)

	tests := []struct {
		name      string
		req       *grpc.AuthRequest
		mockToken string
		mockErr   error
		wantCode  codes.Code
		wantToken string
	}{
		{
			name:      "Successful registration",
			req:       &grpc.AuthRequest{Username: "user1", Password: "pass1"},
			mockToken: "token123",
			wantCode:  codes.OK,
			wantToken: "token123",
		},
		{
			name:     "Invalid username",
			req:      &grpc.AuthRequest{Username: "baduser", Password: "pass1"},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "Invalid password",
			req:      &grpc.AuthRequest{Username: "user1", Password: "badpass"},
			wantCode: codes.InvalidArgument,
		},
		{
			name:     "User already exists",
			req:      &grpc.AuthRequest{Username: "user1", Password: "pass1"},
			mockErr:  services.ErrUserAlreadyExists,
			wantCode: codes.AlreadyExists,
		},
		{
			name:     "Internal error",
			req:      &grpc.AuthRequest{Username: "user1", Password: "pass1"},
			mockErr:  errors.New("internal"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockToken != "" || tt.mockErr != nil {
				mockSvc.EXPECT().
					Register(gomock.Any(), gomock.AssignableToTypeOf(&models.User{})).
					Return(tt.mockToken, tt.mockErr).
					Times(1)
			} else {
				mockSvc.EXPECT().
					Register(gomock.Any(), gomock.Any()).
					Times(0)
			}

			resp, err := handler.Register(context.Background(), tt.req)
			if tt.wantCode != codes.OK {
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantToken, resp.GetToken())
			}
		})
	}
}

func TestGRPCHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)
	handler := NewGRPCHandler(mockSvc, nil, nil) // no validators for login

	tests := []struct {
		name      string
		req       *grpc.AuthRequest
		mockToken string
		mockErr   error
		wantCode  codes.Code
		wantToken string
	}{
		{
			name:      "Successful login",
			req:       &grpc.AuthRequest{Username: "user1", Password: "pass1"},
			mockToken: "token123",
			wantCode:  codes.OK,
			wantToken: "token123",
		},
		{
			name:     "Invalid credentials",
			req:      &grpc.AuthRequest{Username: "user1", Password: "wrongpass"},
			mockErr:  services.ErrInvalidData,
			wantCode: codes.Unauthenticated,
		},
		{
			name:     "Internal error",
			req:      &grpc.AuthRequest{Username: "user1", Password: "pass1"},
			mockErr:  errors.New("internal"),
			wantCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockToken != "" || tt.mockErr != nil {
				mockSvc.EXPECT().
					Authenticate(gomock.Any(), gomock.AssignableToTypeOf(&models.User{})).
					Return(tt.mockToken, tt.mockErr).
					Times(1)
			} else {
				mockSvc.EXPECT().
					Authenticate(gomock.Any(), gomock.Any()).
					Times(0)
			}

			resp, err := handler.Login(context.Background(), tt.req)
			if tt.wantCode != codes.OK {
				assert.Nil(t, resp)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantToken, resp.GetToken())
			}
		})
	}
}
