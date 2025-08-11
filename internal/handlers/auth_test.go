package handlers

import (
	"bytes"
	"encoding/json"
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

	mockSvc := NewMockUserService(ctrl)
	usernameValidator := func(username string) error { return nil }
	passwordValidator := func(password string) error { return nil }
	handler := NewAuthHTTPHandler(mockSvc, usernameValidator, passwordValidator)

	reqBody := RegisterRequest{
		Username: "testuser",
		Password: "testpass",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockSvc.EXPECT().
		Register(gomock.Any(), reqBody.Username, reqBody.Password).
		Return("mocked-token", nil)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer mocked-token", resp.Header.Get("Authorization"))
}

func TestAuthHTTPHandler_Register_InvalidUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)
	usernameValidator := func(username string) error { return errors.New("invalid username") }
	passwordValidator := func(password string) error { return nil }
	handler := NewAuthHTTPHandler(mockSvc, usernameValidator, passwordValidator)

	reqBody := RegisterRequest{
		Username: "baduser",
		Password: "testpass",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthHTTPHandler_Register_UserAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)
	usernameValidator := func(username string) error { return nil }
	passwordValidator := func(password string) error { return nil }
	handler := NewAuthHTTPHandler(mockSvc, usernameValidator, passwordValidator)

	reqBody := RegisterRequest{
		Username: "existinguser",
		Password: "testpass",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockSvc.EXPECT().
		Register(gomock.Any(), reqBody.Username, reqBody.Password).
		Return("", services.ErrUserAlreadyExists)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestAuthHTTPHandler_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)
	handler := NewAuthHTTPHandler(mockSvc, nil, nil) // validators not needed for login

	reqBody := LoginRequest{
		Username: "testuser",
		Password: "testpass",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockSvc.EXPECT().
		Authenticate(gomock.Any(), reqBody.Username, reqBody.Password).
		Return("mocked-token", nil)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer mocked-token", resp.Header.Get("Authorization"))
}

func TestAuthHTTPHandler_Login_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)
	handler := NewAuthHTTPHandler(mockSvc, nil, nil)

	reqBody := LoginRequest{
		Username: "testuser",
		Password: "wrongpass",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockSvc.EXPECT().
		Authenticate(gomock.Any(), reqBody.Username, reqBody.Password).
		Return("", services.ErrInvalidData)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHTTPHandler_Login_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockUserService(ctrl)
	handler := NewAuthHTTPHandler(mockSvc, nil, nil)

	badJSON := []byte(`{"username": "user", "password":`) // malformed JSON

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(badJSON))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
