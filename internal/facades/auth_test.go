package facades

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/gophkeeper/internal/models"
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

	// gomock ожидает вызов GetFromResponse с любым *resty.Response
	mockTokenGetter.EXPECT().GetFromResponse(gomock.Any()).Return("mocked-token", nil)

	req := &models.RegisterRequest{
		Username: "user",
		Password: "pass",
	}

	resp, token, err := facade.Register(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "u123", resp.UserID)
	assert.Equal(t, "d456", resp.DeviceID)
	assert.Equal(t, "mocked-token", token)
	assert.Equal(t, "key123", resp.PrivateKey)
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

	req := &models.LoginRequest{
		Username: "user",
		Password: "pass",
	}

	resp, token, err := facade.Login(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "login-token", token)
	assert.NotNil(t, resp)
}

func TestAuthHTTPFacade_RegisterErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTokenGetter := NewMockTokenGetter(ctrl)

	t.Run("HTTP request error", func(t *testing.T) {
		client := resty.New() // без сервера, запрос вернет ошибку
		facade := NewAuthHTTPFacade(client, mockTokenGetter)

		req := &models.RegisterRequest{
			Username: "user",
			Password: "pass",
		}

		_, _, err := facade.Register(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("HTTP status error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "server error", http.StatusInternalServerError)
		}))
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL)
		facade := NewAuthHTTPFacade(client, mockTokenGetter)

		req := &models.RegisterRequest{
			Username: "user",
			Password: "pass",
		}

		_, _, err := facade.Register(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http error")
	})

	t.Run("TokenGetter error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"user_id":"u1","device_id":"d1","private_key":"k1"}`)
		}))
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL)
		facade := NewAuthHTTPFacade(client, mockTokenGetter)

		mockTokenGetter.EXPECT().GetFromResponse(gomock.Any()).Return("", errors.New("token error"))

		req := &models.RegisterRequest{
			Username: "user",
			Password: "pass",
		}

		_, _, err := facade.Register(context.Background(), req)
		assert.Error(t, err)
		assert.EqualError(t, err, "token error")
	})
}

func TestAuthHTTPFacade_LoginErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockTokenGetter := NewMockTokenGetter(ctrl)

	t.Run("HTTP request error", func(t *testing.T) {
		client := resty.New() // без сервера, запрос вернет ошибку
		facade := NewAuthHTTPFacade(client, mockTokenGetter)

		req := &models.LoginRequest{
			Username: "user",
			Password: "pass",
		}

		_, _, err := facade.Login(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("HTTP status error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "login failed", http.StatusUnauthorized)
		}))
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL)
		facade := NewAuthHTTPFacade(client, mockTokenGetter)

		req := &models.LoginRequest{
			Username: "user",
			Password: "pass",
		}

		_, _, err := facade.Login(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http error")
	})

	t.Run("TokenGetter error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{}`)
		}))
		defer ts.Close()

		client := resty.New().SetBaseURL(ts.URL)
		facade := NewAuthHTTPFacade(client, mockTokenGetter)

		mockTokenGetter.EXPECT().GetFromResponse(gomock.Any()).Return("", errors.New("token error"))

		req := &models.LoginRequest{
			Username: "user",
			Password: "pass",
		}

		_, _, err := facade.Login(context.Background(), req)
		assert.Error(t, err)
		assert.EqualError(t, err, "token error")
	})
}
