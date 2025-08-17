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
	Generate(userID string, deviceID string) (tokenString string, err error)
	SetHeader(w http.ResponseWriter, token string)
}

// UserReader возвращает пользователя по имени.
type UserReader interface {
	Get(ctx context.Context, username string) (*models.UserDB, error)
}

// UserWriter сохраняет нового пользователя.
type UserWriter interface {
	Save(ctx context.Context, userID, username, passwordHash string) error
}

// DeviceWriter сохраняет устройство и связывает его с публичным ключом.
type DeviceWriter interface {
	Save(ctx context.Context, userID, deviceID, publicKey string) error
}

type DeviceReader interface {
	Get(ctx context.Context, userID, deviceID string) (*models.DeviceDB, error)
}

// @Summary      Регистрация нового пользователя
// @Description  Создает новый аккаунт пользователя, генерирует пару ключей RSA и возвращает приватный ключ и ID устройства
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        registerRequest body handlers.RegisterRequest true "Запрос на регистрацию пользователя"
// @Success      200 {object} handlers.RegisterResponse "Успешная регистрация"
// @Failure      400 "Неверный запрос"
// @Failure      409 "Пользователь с таким именем уже существует"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /register [post]
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

		// Хэшируем пароль через bcrypt
		passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		passwordHash := string(passwordHashBytes)

		userID := uuid.New().String()
		deviceID := uuid.New().String()

		// Генерация RSA ключей
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Кодирование приватного ключа
		privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes})

		// Кодирование публичного ключа
		pubBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

		if err := userWriter.Save(ctx, userID, req.Username, passwordHash); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := deviceWriter.Save(ctx, userID, deviceID, string(pubPEM)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		token, err := tokener.Generate(userID, deviceID)
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
// @Description  Проверяет логин и пароль, опционально устройство, генерирует токен и возвращает его в заголовке
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        loginRequest body handlers.LoginRequest true "Запрос на аутентификацию пользователя"
// @Success      200 "Успешная аутентификация, токен в заголовке"
// @Failure      400 "Неверный запрос или устройство не найдено"
// @Failure      401 "Неверный логин или пароль"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /login [post]
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
		if err != nil || user == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Проверка пароля через bcrypt
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

		token, err := tokener.Generate(user.UserID, device.DeviceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tokener.SetHeader(w, token)
	}
}
