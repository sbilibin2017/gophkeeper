package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthHTTPClient_Register_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenGetter := NewMockTokenGetter(ctrl)
	mockTokenGetter.EXPECT().GetFromResponse(gomock.Any()).Return("token123", nil)

	// Spin up a real test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/register", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.RegisterResponse{UserID: "user123"})
	}))
	defer ts.Close()

	authClient := NewAuthHTTPClient(
		resty.New().SetBaseURL(ts.URL),
		mockTokenGetter,
	)

	resp, err := authClient.Register(context.Background(), &models.RegisterRequest{
		Username: "user",
		Password: "pass",
	})

	assert.NoError(t, err)
	assert.Equal(t, "user123", resp.UserID)
	assert.Equal(t, "token123", resp.Token)
}

func TestAuthHTTPClient_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenGetter := NewMockTokenGetter(ctrl)
	mockTokenGetter.EXPECT().GetFromResponse(gomock.Any()).Return("token123", nil)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/login", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		// Can send empty body; token is from TokenGetter
	}))
	defer ts.Close()

	authClient := NewAuthHTTPClient(
		resty.New().SetBaseURL(ts.URL),
		mockTokenGetter,
	)

	resp, err := authClient.Login(context.Background(), &models.LoginRequest{
		Username: "user",
		Password: "pass",
	})

	assert.NoError(t, err)
	assert.Equal(t, "token123", resp.Token)
}

func TestAuthHTTPClient_Login_TokenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTokenGetter := NewMockTokenGetter(ctrl)
	mockTokenGetter.EXPECT().GetFromResponse(gomock.Any()).Return("", assert.AnError)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/login", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	authClient := NewAuthHTTPClient(
		resty.New().SetBaseURL(ts.URL),
		mockTokenGetter,
	)

	resp, err := authClient.Login(context.Background(), &models.LoginRequest{
		Username: "user",
		Password: "pass",
	})

	assert.Nil(t, resp)
	assert.Error(t, err)
}
