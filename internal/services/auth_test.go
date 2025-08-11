package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	authSvc := NewAuthService(mockGetter, mockSaver, mockHasher, mockTokener)

	tests := []struct {
		name       string
		user       *models.User
		setupMocks func()
		wantToken  string
		wantErr    error
	}{
		{
			name: "user_already_exists",
			user: &models.User{Username: "user1", Password: "pass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user1").
					Return(&models.UserDB{Username: "user1"}, nil)
			},
			wantToken: "",
			wantErr:   ErrUserAlreadyExists,
		},
		{
			name: "getter_error",
			user: &models.User{Username: "user2", Password: "pass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user2").
					Return(nil, errors.New("db error"))
			},
			wantToken: "",
			wantErr:   errors.New("db error"),
		},
		{
			name: "hash_error",
			user: &models.User{Username: "user3", Password: "pass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user3").
					Return(nil, nil)
				mockHasher.EXPECT().
					Hash([]byte("pass")).
					Return(nil, errors.New("hash error"))
			},
			wantToken: "",
			wantErr:   errors.New("hash error"),
		},
		{
			name: "save_error",
			user: &models.User{Username: "user4", Password: "pass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user4").
					Return(nil, nil)
				mockHasher.EXPECT().
					Hash([]byte("pass")).
					Return([]byte("hashedpass"), nil)
				mockSaver.EXPECT().
					Save(gomock.Any(), &models.UserDB{
						Username:     "user4",
						PasswordHash: "hashedpass",
					}).
					Return(errors.New("save error"))
			},
			wantToken: "",
			wantErr:   errors.New("save error"),
		},
		{
			name: "token_generate_error",
			user: &models.User{Username: "user5", Password: "pass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user5").
					Return(nil, nil)
				mockHasher.EXPECT().
					Hash([]byte("pass")).
					Return([]byte("hashedpass"), nil)
				mockSaver.EXPECT().
					Save(gomock.Any(), &models.UserDB{
						Username:     "user5",
						PasswordHash: "hashedpass",
					}).
					Return(nil)
				mockTokener.EXPECT().
					Generate("user5").
					Return("", errors.New("token generation failed"))
			},
			wantToken: "",
			wantErr:   errors.New("token generation failed"),
		},
		{
			name: "success",
			user: &models.User{Username: "user6", Password: "pass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user6").
					Return(nil, nil)
				mockHasher.EXPECT().
					Hash([]byte("pass")).
					Return([]byte("hashedpass"), nil)
				mockSaver.EXPECT().
					Save(gomock.Any(), &models.UserDB{
						Username:     "user6",
						PasswordHash: "hashedpass",
					}).
					Return(nil)
				mockTokener.EXPECT().
					Generate("user6").
					Return("token123", nil)
			},
			wantToken: "token123",
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			token, err := authSvc.Register(context.Background(), tt.user)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != tt.wantToken {
				t.Fatalf("expected token %q, got %q", tt.wantToken, token)
			}
		})
	}
}

func TestAuthService_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	authSvc := NewAuthService(mockGetter, nil, mockHasher, mockTokener)

	tests := []struct {
		name       string
		user       *models.User
		setupMocks func()
		wantToken  string
		wantErr    error
	}{
		{
			name:    "user_nil",
			user:    nil,
			wantErr: ErrInvalidData,
		},
		{
			name: "getter_error",
			user: &models.User{Username: "user1", Password: "pass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user1").
					Return(nil, errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "user_not_found",
			user: &models.User{Username: "user2", Password: "pass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user2").
					Return(nil, nil)
			},
			wantErr: ErrInvalidData,
		},
		{
			name: "password_mismatch",
			user: &models.User{Username: "user3", Password: "wrongpass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user3").
					Return(&models.UserDB{Username: "user3", PasswordHash: "hashedpass"}, nil)
				mockHasher.EXPECT().
					Compare([]byte("hashedpass"), []byte("wrongpass")).
					Return(errors.New("password mismatch"))
			},
			wantErr: ErrInvalidData,
		},
		{
			name: "token_generation_error",
			user: &models.User{Username: "user4", Password: "correctpass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user4").
					Return(&models.UserDB{Username: "user4", PasswordHash: "hashedpass"}, nil)
				mockHasher.EXPECT().
					Compare([]byte("hashedpass"), []byte("correctpass")).
					Return(nil)
				mockTokener.EXPECT().
					Generate("user4").
					Return("", errors.New("token error"))
			},
			wantErr: errors.New("token error"),
		},
		{
			name: "success",
			user: &models.User{Username: "user5", Password: "correctpass"},
			setupMocks: func() {
				mockGetter.EXPECT().
					Get(gomock.Any(), "user5").
					Return(&models.UserDB{Username: "user5", PasswordHash: "hashedpass"}, nil)
				mockHasher.EXPECT().
					Compare([]byte("hashedpass"), []byte("correctpass")).
					Return(nil)
				mockTokener.EXPECT().
					Generate("user5").
					Return("token123", nil)
			},
			wantToken: "token123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			token, err := authSvc.Authenticate(context.Background(), tt.user)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err.Error() != tt.wantErr.Error() {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != tt.wantToken {
				t.Errorf("expected token %q, got %q", tt.wantToken, token)
			}
		})
	}
}
