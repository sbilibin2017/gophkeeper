package facades

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthHTTPFacade_RegisterWithMockTokenGetter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenGetter := NewMockTokenGetter(ctrl)

	// Создаем тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer mocked-token")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"user_id":"u123","device_id":"d456","private_key":"key123"}`)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewAuthHTTPFacade(client, mockTokenGetter)

	mockTokenGetter.EXPECT().GetFromResponse(gomock.Any()).Return("mocked-token", nil)

	userID, deviceID, privateKey, token, err := facade.Register(context.Background(), "user", "pass")
	assert.NoError(t, err)
	assert.Equal(t, "u123", userID)
	assert.Equal(t, "d456", deviceID)
	assert.Equal(t, "key123", privateKey)
	assert.Equal(t, "mocked-token", token)
}

func TestAuthHTTPFacade_LoginWithMockTokenGetter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenGetter := NewMockTokenGetter(ctrl)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer login-token")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{}`) // тело ответа не обязательно для Login
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	facade := NewAuthHTTPFacade(client, mockTokenGetter)

	mockTokenGetter.EXPECT().GetFromResponse(gomock.Any()).Return("login-token", nil)

	token, err := facade.Login(context.Background(), "user", "pass")
	assert.NoError(t, err)
	assert.Equal(t, "login-token", token)
}
