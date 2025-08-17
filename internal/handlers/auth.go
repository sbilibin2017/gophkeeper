package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sbilibin2017/gophkeeper/internal/models"
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

// RSAGenerator генерирует пару ключей RSA (приватный и публичный) в формате PEM.
type RSAGenerator interface {
	GenerateKeyPair() (privatePEM string, publicPEM string, err error)
}

type PasswordHasher interface {
	Hash(password string) ([]byte, error)
}

// RegisterRequest определяет входящий запрос на регистрацию пользователя.
// swagger:model RegisterRequest
type RegisterRequest struct {
	// Имя пользователя
	// required: true
	Username string `json:"username"`
	// Пароль пользователя
	// required: true
	Password string `json:"password"`
}

// RegisterResponse определяет ответ на регистрацию пользователя.
// swagger:model RegisterResponse
type RegisterResponse struct {
	// Уникальный идентификатор пользователя
	// required: true
	UserID string `json:"user_id"`
	// Уникальный идентификатор устройства
	// required: true
	DeviceID string `json:"device_id"`
	// Приватный ключ RSA (PEM кодирование)
	// required: true
	PrivateKey string `json:"private_key"`
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
	rsaGenerator RSAGenerator,
	pwHasher PasswordHasher,
	validateUsername func(username string) error,
	validatePassword func(password string) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем контекст запроса
		ctx := r.Context()

		// Декодируем JSON-запрос в структуру RegisterRequest
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Валидация
		err := validateUsername(req.Username)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = validatePassword(req.Password)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Проверяем, существует ли пользователь с таким именем
		if existing, _ := userReader.Get(ctx, req.Username); existing != nil {
			w.WriteHeader(http.StatusConflict) // Пользователь уже существует
			return
		}

		// Генерируем хэш пароля с помощью bcrypt
		passwordHash, err := pwHasher.Hash(req.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Генерируем уникальные идентификаторы для пользователя и устройства
		userID := uuid.New().String()
		deviceID := uuid.New().String()

		// Генерируем пару ключей RSA в формате PEM
		privateKeyPEM, publicKeyPEM, err := rsaGenerator.GenerateKeyPair()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Сохраняем нового пользователя в базе
		if err := userWriter.Save(ctx, userID, req.Username, string(passwordHash)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Сохраняем устройство с публичным ключом
		if err := deviceWriter.Save(ctx, userID, deviceID, publicKeyPEM); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Генерируем токен для нового пользователя и устройства
		token, err := tokener.Generate(userID, deviceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Устанавливаем токен в заголовок HTTP-ответа
		tokener.SetHeader(w, token)
		w.Header().Set("Content-Type", "application/json")

		// Формируем и отправляем JSON-ответ с данными регистрации
		resp := RegisterResponse{
			UserID:     userID,
			DeviceID:   deviceID,
			PrivateKey: privateKeyPEM,
		}
		json.NewEncoder(w).Encode(resp)
	}
}
