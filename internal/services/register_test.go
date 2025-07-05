package services

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/go-yandex-practicum-gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

// -----------------------
// RegisterService (with mock)
// -----------------------

func TestRegisterService_Register(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRegisterer := NewMockRegisterer(ctrl)
		svc := NewRegisterService()
		svc.SetContext(mockRegisterer)

		creds := &models.Credentials{
			Username: "testuser",
			Password: "testpass",
		}

		mockRegisterer.
			EXPECT().
			Register(gomock.Any(), creds).
			Return(nil)

		err := svc.Register(context.Background(), creds)
		assert.NoError(t, err)
	})

	t.Run("no context set", func(t *testing.T) {
		svc := NewRegisterService()
		creds := &models.Credentials{
			Username: "testuser",
			Password: "testpass",
		}

		err := svc.Register(context.Background(), creds)
		assert.EqualError(t, err, "no context set")
	})
}

// -----------------------
// RegisterHTTPService
// -----------------------

func TestRegisterHTTPService_Register(t *testing.T) {
	tests := []struct {
		name    string
		opts    []RegisterHTTPServiceOption
		creds   *models.Credentials
		wantErr bool
	}{
		{
			name: "valid config",
			opts: []RegisterHTTPServiceOption{
				WithHTTPServerURL("http://localhost"),
				WithHTTPPublicKeyPath("/tmp/key.pub"),
				WithHTTPHMACKey("hmac-key-123"),
			},
			creds: &models.Credentials{
				Username: "httpuser",
				Password: "httppass",
			},
			wantErr: false,
		},
		{
			name: "empty config still succeeds (placeholder logic)",
			opts: []RegisterHTTPServiceOption{},
			creds: &models.Credentials{
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewRegisterHTTPService(tt.opts...)
			err := svc.Register(context.Background(), tt.creds)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// -----------------------
// RegisterGRPCService
// -----------------------

func TestRegisterGRPCService_Register(t *testing.T) {
	tests := []struct {
		name    string
		opts    []RegisterGRPCServiceOption
		creds   *models.Credentials
		wantErr bool
	}{
		{
			name: "valid config",
			opts: []RegisterGRPCServiceOption{
				WithGRPCServerURL("localhost:50051"),
				WithGRPCPublicKeyPath("/tmp/key.pub"),
				WithGRPCHMACKey("grpc-hmac-456"),
			},
			creds: &models.Credentials{
				Username: "grpcuser",
				Password: "grpcpass",
			},
			wantErr: false,
		},
		{
			name: "empty config still succeeds (placeholder logic)",
			opts: []RegisterGRPCServiceOption{},
			creds: &models.Credentials{
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewRegisterGRPCService(tt.opts...)
			err := svc.Register(context.Background(), tt.creds)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
