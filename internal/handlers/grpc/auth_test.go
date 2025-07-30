package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthServer_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := NewMockAuthService(ctrl)
	mockJWTGen := NewMockJWTGenerator(ctrl)

	srv := NewAuthServer(mockAuthService, mockJWTGen)

	tests := []struct {
		name        string
		username    string
		password    string
		registerErr error
		jwtToken    string
		jwtGenErr   error
		wantErrCode codes.Code
		wantToken   string
	}{
		{
			name:        "successful register",
			username:    "user1",
			password:    "pass1",
			registerErr: nil,
			jwtToken:    "token123",
			jwtGenErr:   nil,
			wantErrCode: codes.OK,
			wantToken:   "token123",
		},
		{
			name:        "user already exists",
			username:    "user1",
			password:    "pass1",
			registerErr: services.ErrUserAlreadyExists,
			wantErrCode: codes.AlreadyExists,
		},
		{
			name:        "internal register error",
			username:    "user1",
			password:    "pass1",
			registerErr: errors.New("db error"),
			wantErrCode: codes.Internal,
		},
		{
			name:        "jwt generate error",
			username:    "user1",
			password:    "pass1",
			registerErr: nil,
			jwtGenErr:   errors.New("jwt error"),
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService.EXPECT().
				Register(gomock.Any(), tt.username, tt.password).
				Return(tt.registerErr).
				Times(1)

			if tt.registerErr == nil {
				mockJWTGen.EXPECT().
					Generate(tt.username).
					Return(tt.jwtToken, tt.jwtGenErr).
					Times(1)
			}

			req := &pb.AuthRequest{Username: tt.username, Password: tt.password}
			resp, err := srv.Register(context.Background(), req)

			if tt.wantErrCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, resp.GetToken())
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
			}
		})
	}
}

func TestAuthServer_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := NewMockAuthService(ctrl)
	mockJWTGen := NewMockJWTGenerator(ctrl)

	srv := NewAuthServer(mockAuthService, mockJWTGen)

	tests := []struct {
		name        string
		username    string
		password    string
		authErr     error
		jwtToken    string
		jwtGenErr   error
		wantErrCode codes.Code
		wantToken   string
	}{
		{
			name:        "successful login",
			username:    "user1",
			password:    "pass1",
			authErr:     nil,
			jwtToken:    "token123",
			jwtGenErr:   nil,
			wantErrCode: codes.OK,
			wantToken:   "token123",
		},
		{
			name:        "invalid credentials",
			username:    "user1",
			password:    "wrongpass",
			authErr:     services.ErrInvalidData,
			wantErrCode: codes.Unauthenticated,
		},
		{
			name:        "internal auth error",
			username:    "user1",
			password:    "pass1",
			authErr:     errors.New("db error"),
			wantErrCode: codes.Internal,
		},
		{
			name:        "jwt generate error",
			username:    "user1",
			password:    "pass1",
			authErr:     nil,
			jwtGenErr:   errors.New("jwt error"),
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthService.EXPECT().
				Authenticate(gomock.Any(), tt.username, tt.password).
				Return(tt.authErr).
				Times(1)

			if tt.authErr == nil {
				mockJWTGen.EXPECT().
					Generate(tt.username).
					Return(tt.jwtToken, tt.jwtGenErr).
					Times(1)
			}

			req := &pb.AuthRequest{Username: tt.username, Password: tt.password}
			resp, err := srv.Login(context.Background(), req)

			if tt.wantErrCode == codes.OK {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, resp.GetToken())
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantErrCode, st.Code())
			}
		})
	}
}
