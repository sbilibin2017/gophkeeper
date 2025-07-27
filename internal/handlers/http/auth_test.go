package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to encode JSON
func toJSON(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func TestRegisterHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)

	reqBody := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}

	token := "mock-token"
	mockReg.
		EXPECT().
		Register(gomock.Any(), "testuser", "testpass").
		Return(&token, nil)

	handler := NewRegisterHandler(mockReg)

	req := httptest.NewRequest(http.MethodPost, "/register", toJSON(t, reqBody))
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer "+token, resp.Header.Get("Authorization"))
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)
	handler := NewRegisterHandler(mockReg)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not-json"))
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestRegisterHandler_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)

	mockReg.
		EXPECT().
		Register(gomock.Any(), "user", "pass").
		Return(nil, assert.AnError)

	handler := NewRegisterHandler(mockReg)

	body := toJSON(t, map[string]string{"username": "user", "password": "pass"})
	req := httptest.NewRequest(http.MethodPost, "/register", body)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestLoginHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogin := NewMockLoginer(ctrl)

	token := "login-token"
	mockLogin.
		EXPECT().
		Login(gomock.Any(), "user", "pass").
		Return(&token, nil)

	handler := NewLoginHandler(mockLogin)

	body := toJSON(t, map[string]string{"username": "user", "password": "pass"})
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Bearer "+token, resp.Header.Get("Authorization"))
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogin := NewMockLoginer(ctrl)
	handler := NewLoginHandler(mockLogin)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("bad json"))
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestLoginHandler_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogin := NewMockLoginer(ctrl)

	mockLogin.
		EXPECT().
		Login(gomock.Any(), "user", "wrong").
		Return(nil, assert.AnError)

	handler := NewLoginHandler(mockLogin)

	body := toJSON(t, map[string]string{"username": "user", "password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
