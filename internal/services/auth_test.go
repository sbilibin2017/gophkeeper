package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGetter := NewMockUserGetter(ctrl)
	mockUserSaver := NewMockUserSaver(ctrl)

	service := NewAuthService(mockUserGetter, mockUserSaver)

	tests := []struct {
		name          string
		username      string
		password      string
		getUserReturn *models.User
		getUserErr    error
		saveErr       error
		expectErr     error
	}{
		{
			name:          "success",
			username:      "alice",
			password:      "password123",
			getUserReturn: nil,
			getUserErr:    nil,
			saveErr:       nil,
			expectErr:     nil,
		},
		{
			name:          "user already exists",
			username:      "bob",
			password:      "password123",
			getUserReturn: &models.User{Username: "bob"},
			getUserErr:    nil,
			saveErr:       nil,
			expectErr:     ErrUserAlreadyExists,
		},
		{
			name:          "error getting user",
			username:      "charlie",
			password:      "password123",
			getUserReturn: nil,
			getUserErr:    errors.New("db error"),
			saveErr:       nil,
			expectErr:     errors.New("db error"),
		},
		{
			name:          "error saving user",
			username:      "david",
			password:      "password123",
			getUserReturn: nil,
			getUserErr:    nil,
			saveErr:       errors.New("save failed"),
			expectErr:     errors.New("save failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserGetter.EXPECT().Get(gomock.Any(), tt.username).Return(tt.getUserReturn, tt.getUserErr)
			if tt.getUserReturn == nil && tt.getUserErr == nil {
				mockUserSaver.EXPECT().Save(gomock.Any(), tt.username, gomock.Any()).Return(tt.saveErr)
			}

			err := service.Register(context.Background(), tt.username, tt.password)
			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserGetter := NewMockUserGetter(ctrl)
	service := NewAuthService(mockUserGetter, nil)

	password := "correct_password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name          string
		username      string
		passwordInput string
		getUserReturn *models.User
		getUserErr    error
		expectErr     error
	}{
		{
			name:          "success",
			username:      "alice",
			passwordInput: password,
			getUserReturn: &models.User{Username: "alice", PasswordHash: string(hashedPassword)},
			getUserErr:    nil,
			expectErr:     nil,
		},
		{
			name:          "user not found",
			username:      "bob",
			passwordInput: "somepass",
			getUserReturn: nil,
			getUserErr:    nil,
			expectErr:     ErrInvalidData,
		},
		{
			name:          "error getting user",
			username:      "charlie",
			passwordInput: "somepass",
			getUserReturn: nil,
			getUserErr:    errors.New("db error"),
			expectErr:     ErrInvalidData,
		},
		{
			name:          "wrong password",
			username:      "david",
			passwordInput: "wrongpassword",
			getUserReturn: &models.User{Username: "david", PasswordHash: string(hashedPassword)},
			getUserErr:    nil,
			expectErr:     ErrInvalidData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserGetter.EXPECT().Get(gomock.Any(), tt.username).Return(tt.getUserReturn, tt.getUserErr)

			err := service.Authenticate(context.Background(), tt.username, tt.passwordInput)
			if tt.expectErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
