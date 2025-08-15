package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestAuthHTTPHandler_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockAuthenticator(ctrl)
	handler := NewAuthHTTPHandler(mockSvc, nil, nil)

	reqBody := `{"username":"user1","password":"pass1","device_id":"device1"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	expectedPriv := []byte("private-key")
	expectedToken := "token"

	mockSvc.EXPECT().
		Register(gomock.Any(), "user1", "pass1", "device1").
		Return(expectedPriv, expectedToken, nil)

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer "+expectedToken, resp.Header.Get("Authorization"))
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	assert.Equal(t, expectedPriv, buf.Bytes())
}

func TestAuthHTTPHandler_Register_BadRequest_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockAuthenticator(ctrl)
	handler := NewAuthHTTPHandler(mockSvc, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("{invalid-json}"))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestAuthHTTPHandler_Register_UserExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockAuthenticator(ctrl)
	handler := NewAuthHTTPHandler(mockSvc, nil, nil)

	reqBody := `{"username":"user1","password":"pass1","device_id":"device1"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	mockSvc.EXPECT().
		Register(gomock.Any(), "user1", "pass1", "device1").
		Return(nil, "", services.ErrUserExists)

	handler.Register(w, req)

	assert.Equal(t, http.StatusConflict, w.Result().StatusCode)
}

func TestAuthHTTPHandler_Register_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockAuthenticator(ctrl)
	handler := NewAuthHTTPHandler(mockSvc, nil, nil)

	reqBody := `{"username":"user1","password":"pass1","device_id":"device1"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	mockSvc.EXPECT().
		Register(gomock.Any(), "user1", "pass1", "device1").
		Return(nil, "", errors.New("some error"))

	handler.Register(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)
}

func TestAuthHTTPHandler_Register_ValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockAuthenticator(ctrl)
	usernameValidator := func(username string) error {
		return errors.New("invalid username")
	}
	passwordValidator := func(password string) error {
		return errors.New("invalid password")
	}

	handler := NewAuthHTTPHandler(mockSvc, usernameValidator, passwordValidator)

	reqBody := `{"username":"user1","password":"pass1","device_id":"device1"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestAuthHTTPHandler_Register_PasswordValidationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockAuthenticator(ctrl)

	// Валидатор пароля, который всегда возвращает ошибку
	passwordValidator := func(password string) error {
		return errors.New("invalid password")
	}

	handler := NewAuthHTTPHandler(mockSvc, nil, passwordValidator)

	reqBody := `{"username":"user1","password":"badpass","device_id":"device1"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(reqBody))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Ожидаем 400 Bad Request из-за ошибки валидации пароля
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
