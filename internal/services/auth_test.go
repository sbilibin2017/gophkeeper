package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

func TestAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"
	hashedPassword := []byte("hashedpassword")
	token := "jwt-token"

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	mockGetter.EXPECT().Get(ctx, username).Return(nil, nil)
	mockHasher.EXPECT().Hash([]byte(password)).Return(hashedPassword, nil)
	mockSaver.EXPECT().Save(ctx, username, string(hashedPassword)).Return(nil)
	mockTokener.EXPECT().Generate(username).Return(token, nil)

	authService := NewAuthService(mockGetter, mockSaver, mockHasher, mockTokener)

	gotToken, err := authService.Register(ctx, username, password)
	assert.NoError(t, err)
	assert.Equal(t, token, gotToken)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	mockGetter.EXPECT().Get(ctx, username).Return(&models.UserDB{}, nil)

	authService := NewAuthService(mockGetter, mockSaver, mockHasher, mockTokener)

	token, err := authService.Register(ctx, username, password)
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
	assert.Empty(t, token)
}

func TestAuthService_Register_HashError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	mockGetter.EXPECT().Get(ctx, username).Return(nil, nil)
	mockHasher.EXPECT().Hash([]byte(password)).Return(nil, errors.New("hash error"))

	authService := NewAuthService(mockGetter, mockSaver, mockHasher, mockTokener)

	token, err := authService.Register(ctx, username, password)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAuthService_Register_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"
	hashedPassword := []byte("hashedpassword")

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	mockGetter.EXPECT().Get(ctx, username).Return(nil, nil)
	mockHasher.EXPECT().Hash([]byte(password)).Return(hashedPassword, nil)
	mockSaver.EXPECT().Save(ctx, username, string(hashedPassword)).Return(errors.New("save error"))

	authService := NewAuthService(mockGetter, mockSaver, mockHasher, mockTokener)

	token, err := authService.Register(ctx, username, password)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAuthService_Register_GenerateTokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"
	hashedPassword := []byte("hashedpassword")

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	mockGetter.EXPECT().Get(ctx, username).Return(nil, nil)
	mockHasher.EXPECT().Hash([]byte(password)).Return(hashedPassword, nil)
	mockSaver.EXPECT().Save(ctx, username, string(hashedPassword)).Return(nil)
	mockTokener.EXPECT().Generate(username).Return("", errors.New("token generation failed"))

	authService := NewAuthService(mockGetter, mockSaver, mockHasher, mockTokener)

	token, err := authService.Register(ctx, username, password)
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAuthService_Authenticate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"
	hashedPassword := []byte("hashedpassword")
	token := "jwt-token"

	mockGetter := NewMockUserGetter(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	userDB := &models.UserDB{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	mockGetter.EXPECT().Get(ctx, username).Return(userDB, nil)
	mockHasher.EXPECT().Compare(hashedPassword, []byte(password)).Return(nil)
	mockTokener.EXPECT().Generate(username).Return(token, nil)

	authService := NewAuthService(mockGetter, nil, mockHasher, mockTokener)

	gotToken, err := authService.Authenticate(ctx, username, password)
	assert.NoError(t, err)
	assert.Equal(t, token, gotToken)
}

func TestAuthService_Authenticate_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"

	mockGetter := NewMockUserGetter(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	mockGetter.EXPECT().Get(ctx, username).Return(nil, nil)

	authService := NewAuthService(mockGetter, nil, mockHasher, mockTokener)

	token, err := authService.Authenticate(ctx, username, password)
	assert.ErrorIs(t, err, ErrInvalidData)
	assert.Empty(t, token)
}

func TestAuthService_Authenticate_CompareError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"
	hashedPassword := []byte("hashedpassword")

	mockGetter := NewMockUserGetter(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	userDB := &models.UserDB{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	mockGetter.EXPECT().Get(ctx, username).Return(userDB, nil)
	mockHasher.EXPECT().Compare(hashedPassword, []byte(password)).Return(errors.New("password mismatch"))

	authService := NewAuthService(mockGetter, nil, mockHasher, mockTokener)

	token, err := authService.Authenticate(ctx, username, password)
	assert.ErrorIs(t, err, ErrInvalidData)
	assert.Empty(t, token)
}

func TestAuthService_Authenticate_GenerateTokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "user1"
	password := "password123"
	hashedPassword := []byte("hashedpassword")

	mockGetter := NewMockUserGetter(ctrl)
	mockHasher := NewMockHasher(ctrl)
	mockTokener := NewMockTokener(ctrl)

	userDB := &models.UserDB{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	mockGetter.EXPECT().Get(ctx, username).Return(userDB, nil)
	mockHasher.EXPECT().Compare(hashedPassword, []byte(password)).Return(nil)
	mockTokener.EXPECT().Generate(username).Return("", errors.New("token generation failed"))

	authService := NewAuthService(mockGetter, nil, mockHasher, mockTokener)

	token, err := authService.Authenticate(ctx, username, password)
	assert.Error(t, err)
	assert.Empty(t, token)
}
