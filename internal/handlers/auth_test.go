package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/gophkeeper/internal/handlers"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := handlers.NewMockUserReader(ctrl)
	mockUserWriter := handlers.NewMockUserWriter(ctrl)
	mockDeviceWriter := handlers.NewMockDeviceWriter(ctrl)
	mockTokener := handlers.NewMockTokener(ctrl)

	validateUsername := func(username string) error {
		if username == "" {
			return errors.New("invalid username")
		}
		return nil
	}
	validatePassword := func(password string) error {
		if len(password) < 6 {
			return errors.New("password too short")
		}
		return nil
	}

	handler := handlers.NewRegisterHTTPHandler(
		mockUserReader,
		mockUserWriter,
		mockDeviceWriter,
		mockTokener,
		validateUsername,
		validatePassword,
	)

	t.Run("success", func(t *testing.T) {
		reqBody := models.RegisterRequest{
			Username: "alice",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		mockUserReader.EXPECT().Get(gomock.Any(), "alice").Return(nil, nil)
		mockUserWriter.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
		mockDeviceWriter.EXPECT().Save(gomock.Any(), gomock.Any()).Return(nil)
		mockTokener.EXPECT().Generate(gomock.Any()).Return("token123", nil)
		mockTokener.EXPECT().SetHeader(gomock.Any(), "token123")

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp models.RegisterResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.NotEmpty(t, resp.UserID)
		assert.NotEmpty(t, resp.DeviceID)
		assert.NotEmpty(t, resp.PrivateKey)
	})

	t.Run("username exists", func(t *testing.T) {
		reqBody := models.RegisterRequest{Username: "bob", Password: "pass1234"}
		body, _ := json.Marshal(reqBody)

		mockUserReader.EXPECT().Get(gomock.Any(), "bob").Return(&models.UserDB{}, nil)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)
		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("invalid password", func(t *testing.T) {
		reqBody := models.RegisterRequest{Username: "bob", Password: "123"}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := handlers.NewMockUserReader(ctrl)
	mockDeviceReader := handlers.NewMockDeviceReader(ctrl)
	mockTokener := handlers.NewMockTokener(ctrl)

	handler := handlers.NewLoginHTTPHandler(mockUserReader, mockDeviceReader, mockTokener)

	t.Run("success", func(t *testing.T) {
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		user := &models.UserDB{
			UserID:       "user123",
			Username:     "alice",
			PasswordHash: string(passwordHash),
		}
		device := &models.DeviceDB{
			UserID:   "user123",
			DeviceID: "device123",
		}

		reqBody := models.LoginRequest{Username: "alice", Password: "password123", DeviceID: "device123"}
		body, _ := json.Marshal(reqBody)

		mockUserReader.EXPECT().Get(gomock.Any(), "alice").Return(user, nil)
		mockDeviceReader.EXPECT().Get(gomock.Any(), "user123", "device123").Return(device, nil)
		mockTokener.EXPECT().Generate(gomock.Any()).Return("token123", nil)
		mockTokener.EXPECT().SetHeader(gomock.Any(), "token123")

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid password", func(t *testing.T) {
		passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		user := &models.UserDB{UserID: "user123", Username: "alice", PasswordHash: string(passwordHash)}

		reqBody := models.LoginRequest{Username: "alice", Password: "wrongpass", DeviceID: "device123"}
		body, _ := json.Marshal(reqBody)

		mockUserReader.EXPECT().Get(gomock.Any(), "alice").Return(user, nil)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
