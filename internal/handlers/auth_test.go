package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// Test NewRegisterHandler success case
func TestRegisterHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "testuser"
	password := "pass123"
	token := "jwt-token"

	// UserGetter.Get returns nil (user does not exist)
	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, nil)

	// UserSaver.Save succeeds
	mockSaver.EXPECT().
		Save(gomock.Any(), username, gomock.Any()).
		Return(nil)

	// JWTGenerator.Generate returns token
	mockJWT.EXPECT().
		Generate(username).
		Return(token, nil)

	handler := NewRegisterHandler(mockGetter, mockSaver, mockJWT)

	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Bearer "+token, rec.Header().Get("Authorization"))
}

// Test NewRegisterHandler when user already exists
func TestRegisterHandler_UserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockSaver := NewMockUserSaver(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "existinguser"

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(&models.User{Username: username}, nil)

	handler := NewRegisterHandler(mockGetter, mockSaver, mockJWT)

	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "irrelevant",
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), errUserAlreadyExists.Error())
}

// Test NewLoginHandler success case
func TestLoginHandler_Success(t *testing.T) {
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

	handler := NewLoginHandler(mockGetter, mockJWT)

	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "Bearer "+token, rec.Header().Get("Authorization"))
}

// Test NewLoginHandler invalid password
func TestLoginHandler_InvalidPassword(t *testing.T) {
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

	handler := NewLoginHandler(mockGetter, mockJWT)

	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": wrongPassword,
	})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), errInvalidLogin.Error())
}

// Test NewLoginHandler user not found
func TestLoginHandler_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockGetter := NewMockUserGetter(ctrl)
	mockJWT := NewMockJWTGenerator(ctrl)

	username := "nonexistent"

	mockGetter.EXPECT().
		Get(gomock.Any(), username).
		Return(nil, errors.New("not found"))

	handler := NewLoginHandler(mockGetter, mockJWT)

	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": "any",
	})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBody))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), errInvalidLogin.Error())
}
