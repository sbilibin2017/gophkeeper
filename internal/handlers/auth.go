package handlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"

	"github.com/google/uuid"
	"github.com/sbilibin2017/gophkeeper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Tokener генерирует JWT или другие токены.
type Tokener interface {
	Generate(payload *models.TokenPayload) (string, error)
	SetHeader(w http.ResponseWriter, token string)
}

// UserReader возвращает пользователя по имени.
type UserReader interface {
	Get(ctx context.Context, username string) (*models.UserDB, error)
}

// UserWriter сохраняет нового пользователя.
type UserWriter interface {
	Save(ctx context.Context, user *models.UserDB) error
}

// DeviceWriter сохраняет устройство и связывает его с публичным ключом.
type DeviceWriter interface {
	Save(ctx context.Context, device *models.DeviceDB) error
}

// DeviceReader возвращает устройство по userID и deviceID.
type DeviceReader interface {
	Get(ctx context.Context, userID, deviceID string) (*models.DeviceDB, error)
}

// @Summary      Регистрация нового пользователя
// @Description  Создает новый аккаунт пользователя, генерирует пару ключей RSA, возвращает приватный ключ и ID устройства. Токен JWT возвращается в заголовке.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.RegisterRequest true "Запрос на регистрацию пользователя"
// @Success      200 {object} models.RegisterResponse "Успешная регистрация"
// @Failure      400 {string} string "Неверный запрос или невалидные данные"
// @Failure      409 {string} string "Пользователь с таким именем уже существует"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /auth/register [post]
func NewRegisterHTTPHandler(
	userReader UserReader,
	userWriter UserWriter,
	deviceWriter DeviceWriter,
	tokener Tokener,
	validateUsername func(username string) error,
	validatePassword func(password string) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req models.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := validateUsername(req.Username); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := validatePassword(req.Password); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if existing, _ := userReader.Get(ctx, req.Username); existing != nil {
			w.WriteHeader(http.StatusConflict)
			return
		}

		passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userID := uuid.New().String()
		deviceID := uuid.New().String()

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

		pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

		// Создаём пользователя
		user := &models.UserDB{
			UserID:       userID,
			Username:     req.Username,
			PasswordHash: string(passwordHashBytes),
		}
		if err := userWriter.Save(ctx, user); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Создаём устройство
		device := &models.DeviceDB{
			UserID:    userID,
			DeviceID:  deviceID,
			PublicKey: string(pubPEM),
		}
		if err := deviceWriter.Save(ctx, device); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Генерация токена
		tokenPayload := &models.TokenPayload{
			UserID:   userID,
			DeviceID: deviceID,
		}
		token, err := tokener.Generate(tokenPayload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tokener.SetHeader(w, token)

		w.Header().Set("Content-Type", "application/json")
		resp := models.RegisterResponse{
			UserID:     userID,
			DeviceID:   deviceID,
			PrivateKey: string(privPEM),
		}
		json.NewEncoder(w).Encode(resp)
	}
}

// @Summary      Аутентификация пользователя
// @Description  Проверяет логин и пароль, а также устройство. Генерирует JWT токен и возвращает его в заголовке Authorization.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body models.LoginRequest true "Запрос на аутентификацию пользователя"
// @Success      200 {string} string "Успешная аутентификация, токен в заголовке Authorization"
// @Failure      400 {string} string "Неверный запрос или устройство не найдено"
// @Failure      401 {string} string "Неверный логин или пароль"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /auth/login [post]
func NewLoginHTTPHandler(
	userReader UserReader,
	deviceReader DeviceReader,
	tokener Tokener,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req models.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		user, err := userReader.Get(ctx, req.Username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		device, err := deviceReader.Get(ctx, user.UserID, req.DeviceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if device == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenPayload := &models.TokenPayload{
			UserID:   user.UserID,
			DeviceID: device.DeviceID,
		}
		token, err := tokener.Generate(tokenPayload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tokener.SetHeader(w, token)
		w.WriteHeader(http.StatusOK)
	}
}
