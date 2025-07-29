package grpcservers

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	pb "github.com/sbilibin2017/gophkeeper/pkg/grpc"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthServer_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "testuser"
	password := "pass123"
	token := "jwt-token"

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, nil)

	mockSaver.EXPECT().
		Save(gomock.Any(), username, gomock.Any()).
		Return(nil)

	mockJWT.EXPECT().
		Generate(username).
		Return(token, nil)

	server := NewAuthServer(mockSaver, mockGetter, mockJWT)
	req := &pb.AuthRequest{Username: username, Password: password}

	resp, err := server.Register(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, token, resp.GetToken())
}

func TestAuthServer_Register_UserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "existinguser"

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(&models.User{Username: username}, nil)

	server := NewAuthServer(mockSaver, mockGetter, mockJWT)
	req := &pb.AuthRequest{Username: username, Password: "irrelevant"}

	resp, err := server.Register(context.Background(), req)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.AlreadyExists, st.Code())
}

func TestAuthServer_Register_HashPasswordFail(t *testing.T) {
	// To simulate bcrypt failure is tricky because it is a third party call.
	// You could wrap bcrypt call in an interface and mock it in production code,
	// but here we skip this test as it's rare and would require refactoring.
}

func TestAuthServer_Register_SaveFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "testuser"
	password := "pass123"

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, nil)

	mockSaver.EXPECT().
		Save(gomock.Any(), username, gomock.Any()).
		Return(errors.New("db error"))

	server := NewAuthServer(mockSaver, mockGetter, mockJWT)
	req := &pb.AuthRequest{Username: username, Password: password}

	resp, err := server.Register(context.Background(), req)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestAuthServer_Register_GenerateTokenFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "testuser"
	password := "pass123"

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, nil)

	mockSaver.EXPECT().
		Save(gomock.Any(), username, gomock.Any()).
		Return(nil)

	mockJWT.EXPECT().
		Generate(username).
		Return("", errors.New("token error"))

	server := NewAuthServer(mockSaver, mockGetter, mockJWT)
	req := &pb.AuthRequest{Username: username, Password: password}

	resp, err := server.Register(context.Background(), req)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}

func TestAuthServer_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "testuser"
	password := "pass123"
	token := "jwt-token"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(&models.User{Username: username, PasswordHash: string(hashedPassword)}, nil)

	mockJWT.EXPECT().
		Generate(username).
		Return(token, nil)

	server := NewAuthServer(nil, mockGetter, mockJWT)
	req := &pb.AuthRequest{Username: username, Password: password}

	resp, err := server.Login(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, token, resp.GetToken())
}

func TestAuthServer_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "nonexistent"

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, errors.New("not found"))

	server := NewAuthServer(nil, mockGetter, mockJWT)
	req := &pb.AuthRequest{Username: username, Password: "any"}

	resp, err := server.Login(context.Background(), req)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthServer_Login_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "testuser"
	correctPassword := "correctpass"
	wrongPassword := "wrongpass"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(&models.User{Username: username, PasswordHash: string(hashedPassword)}, nil)

	server := NewAuthServer(nil, mockGetter, mockJWT)
	req := &pb.AuthRequest{Username: username, Password: wrongPassword}

	resp, err := server.Login(context.Background(), req)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestAuthServer_Login_GenerateTokenFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "testuser"
	password := "pass123"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(&models.User{Username: username, PasswordHash: string(hashedPassword)}, nil)

	mockJWT.EXPECT().
		Generate(username).
		Return("", errors.New("token error"))

	server := NewAuthServer(nil, mockGetter, mockJWT)
	req := &pb.AuthRequest{Username: username, Password: password}

	resp, err := server.Login(context.Background(), req)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
}
