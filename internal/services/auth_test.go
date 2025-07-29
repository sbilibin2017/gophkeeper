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

func TestAuthService_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockUserReader(ctrl)
	mockWriter := NewMockUserWriter(ctrl)

	username := "newuser"
	password := "securepass"

	mockReader.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, nil)

	mockWriter.EXPECT().
		Save(gomock.Any(), username, gomock.Any()).
		Return(nil)

	auth := &AuthService{
		reader: mockReader,
		writer: mockWriter,
	}

	err := auth.Register(context.Background(), username, password)
	assert.NoError(t, err)
}

func TestAuthService_Register_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockUserReader(ctrl)
	mockWriter := NewMockUserWriter(ctrl)

	username := "existinguser"

	mockReader.EXPECT().
		Get(gomock.Any(), username).
		Return(&models.User{Username: username}, nil)

	auth := &AuthService{
		reader: mockReader,
		writer: mockWriter,
	}

	err := auth.Register(context.Background(), username, "any")
	assert.ErrorIs(t, err, ErrUserAlreadyExists)
}

func TestAuthService_Register_HashError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockUserReader(ctrl)
	mockWriter := NewMockUserWriter(ctrl)

	username := "failhash"
	// Intentionally trigger bcrypt error by passing absurd length
	password := string(make([]byte, 1<<20))

	mockReader.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, nil)

	auth := &AuthService{
		reader: mockReader,
		writer: mockWriter,
	}

	err := auth.Register(context.Background(), username, password)
	assert.Error(t, err)
}

func TestAuthService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockUserReader(ctrl)

	username := "validuser"
	password := "validpassword"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockReader.EXPECT().
		Get(gomock.Any(), username).
		Return(&models.User{
			Username:     username,
			PasswordHash: string(hashed),
		}, nil)

	auth := &AuthService{
		reader: mockReader,
	}

	err := auth.Login(context.Background(), username, password)
	assert.NoError(t, err)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockUserReader(ctrl)

	username := "user"
	password := "correctpass"
	wrongPassword := "wrongpass"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockReader.EXPECT().
		Get(gomock.Any(), username).
		Return(&models.User{
			Username:     username,
			PasswordHash: string(hashed),
		}, nil)

	auth := &AuthService{
		reader: mockReader,
	}

	err := auth.Login(context.Background(), username, wrongPassword)
	assert.EqualError(t, err, "invalid username or password")
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReader := NewMockUserReader(ctrl)

	username := "unknown"

	mockReader.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, errors.New("not found"))

	auth := &AuthService{
		reader: mockReader,
	}

	err := auth.Login(context.Background(), username, "somepassword")
	assert.EqualError(t, err, "not found")
}
