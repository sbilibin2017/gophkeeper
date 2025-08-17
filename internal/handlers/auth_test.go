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
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockUserReader(ctrl)
	mockUserWriter := NewMockUserWriter(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)
	mockTokener := NewMockTokener(ctrl)
	mockRSAGenerator := NewMockRSAGenerator(ctrl)
	mockPasswordHasher := NewMockPasswordHasher(ctrl)

	validateUsername := func(username string) error {
		if username == "" {
			return errors.New("username required")
		}
		return nil
	}
	validatePassword := func(password string) error {
		if password == "" {
			return errors.New("password required")
		}
		return nil
	}

	handler := NewRegisterHTTPHandler(
		mockUserReader,
		mockUserWriter,
		mockDeviceWriter,
		mockTokener,
		mockRSAGenerator,
		mockPasswordHasher,
		validateUsername,
		validatePassword,
	)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := RegisterRequest{
			Username: "testuser",
			Password: "securepassword",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mockUserReader.EXPECT().Get(gomock.Any(), "testuser").Return(nil, nil)
		mockRSAGenerator.EXPECT().GenerateKeyPair().Return("privatePEM", "publicPEM", nil)
		mockPasswordHasher.EXPECT().Hash("securepassword").Return([]byte("hashedPassword"), nil)
		mockUserWriter.EXPECT().Save(gomock.Any(), gomock.Any(), "testuser", "hashedPassword").Return(nil)
		mockDeviceWriter.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any(), "publicPEM").Return(nil)
		mockTokener.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("token123", nil)
		mockTokener.EXPECT().SetHeader(w, "token123")

		handler(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var registerResp RegisterResponse
		err := json.NewDecoder(resp.Body).Decode(&registerResp)
		assert.NoError(t, err)
		assert.Equal(t, "privatePEM", registerResp.PrivateKey)
		assert.NotEmpty(t, registerResp.UserID)
		assert.NotEmpty(t, registerResp.DeviceID)
	})

	t.Run("username already exists", func(t *testing.T) {
		reqBody := RegisterRequest{
			Username: "existinguser",
			Password: "password",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		mockUserReader.EXPECT().Get(gomock.Any(), "existinguser").Return(&models.UserDB{}, nil)

		handler(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("invalid username", func(t *testing.T) {
		reqBody := RegisterRequest{
			Username: "",
			Password: "password",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid password", func(t *testing.T) {
		reqBody := RegisterRequest{
			Username: "validuser",
			Password: "",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, req)
		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestRegisterHandler_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserReader := NewMockUserReader(ctrl)
	mockUserWriter := NewMockUserWriter(ctrl)
	mockDeviceWriter := NewMockDeviceWriter(ctrl)
	mockTokener := NewMockTokener(ctrl)
	mockRSA := NewMockRSAGenerator(ctrl)
	mockHasher := NewMockPasswordHasher(ctrl)

	validateUsername := func(username string) error {
		if username == "baduser" {
			return errors.New("invalid username")
		}
		return nil
	}
	validatePassword := func(password string) error {
		if password == "badpass" {
			return errors.New("invalid password")
		}
		return nil
	}

	handler := NewRegisterHTTPHandler(
		mockUserReader, mockUserWriter, mockDeviceWriter,
		mockTokener, mockRSA, mockHasher,
		validateUsername, validatePassword,
	)

	tests := []struct {
		name           string
		body           string
		setupMocks     func()
		expectedStatus int
	}{
		{
			name:           "invalid JSON",
			body:           `{bad json}`,
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid username",
			body:           `{"username":"baduser","password":"goodpass"}`,
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid password",
			body:           `{"username":"gooduser","password":"badpass"}`,
			setupMocks:     func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user already exists",
			body: `{"username":"existinguser","password":"goodpass"}`,
			setupMocks: func() {
				mockUserReader.EXPECT().Get(gomock.Any(), "existinguser").
					Return(&models.UserDB{}, nil)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "hash error",
			body: `{"username":"newuser","password":"goodpass"}`,
			setupMocks: func() {
				mockUserReader.EXPECT().Get(gomock.Any(), "newuser").Return(nil, nil)
				mockHasher.EXPECT().Hash("goodpass").Return(nil, errors.New("hash error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "RSA generation error",
			body: `{"username":"newuser","password":"goodpass"}`,
			setupMocks: func() {
				mockUserReader.EXPECT().Get(gomock.Any(), "newuser").Return(nil, nil)
				mockHasher.EXPECT().Hash("goodpass").Return([]byte("hashed"), nil)
				mockRSA.EXPECT().GenerateKeyPair().Return("", "", errors.New("rsa error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "user save error",
			body: `{"username":"newuser","password":"goodpass"}`,
			setupMocks: func() {
				mockUserReader.EXPECT().Get(gomock.Any(), "newuser").Return(nil, nil)
				mockHasher.EXPECT().Hash("goodpass").Return([]byte("hashed"), nil)
				mockRSA.EXPECT().GenerateKeyPair().Return("priv", "pub", nil)
				mockUserWriter.EXPECT().Save(gomock.Any(), gomock.Any(), "newuser", "hashed").Return(errors.New("save error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "device save error",
			body: `{"username":"newuser","password":"goodpass"}`,
			setupMocks: func() {
				mockUserReader.EXPECT().Get(gomock.Any(), "newuser").Return(nil, nil)
				mockHasher.EXPECT().Hash("goodpass").Return([]byte("hashed"), nil)
				mockRSA.EXPECT().GenerateKeyPair().Return("priv", "pub", nil)
				mockUserWriter.EXPECT().Save(gomock.Any(), gomock.Any(), "newuser", "hashed").Return(nil)
				mockDeviceWriter.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any(), "pub").
					Return(errors.New("device save error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "token generation error",
			body: `{"username":"newuser","password":"goodpass"}`,
			setupMocks: func() {
				mockUserReader.EXPECT().Get(gomock.Any(), "newuser").Return(nil, nil)
				mockHasher.EXPECT().Hash("goodpass").Return([]byte("hashed"), nil)
				mockRSA.EXPECT().GenerateKeyPair().Return("priv", "pub", nil)
				mockUserWriter.EXPECT().Save(gomock.Any(), gomock.Any(), "newuser", "hashed").Return(nil)
				mockDeviceWriter.EXPECT().Save(gomock.Any(), gomock.Any(), gomock.Any(), "pub").Return(nil)
				mockTokener.EXPECT().Generate(gomock.Any(), gomock.Any()).Return("", errors.New("token error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tt.body))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}
