package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/sbilibin2017/gophkeeper/internal/services"
)

// Authenticator описывает интерфейс для регистрации пользователей.
type Authenticator interface {
	// Register создаёт нового пользователя и устройство.
	// Возвращает приватный ключ устройства и токен авторизации.
	Register(
		ctx context.Context,
		username string,
		password string,
		deviceID string,
	) ([]byte, string, error)
}

// AuthRequest представляет тело запроса на регистрацию.
//
// swagger:model AuthRequest
type AuthRequest struct {
	// Username имя пользователя
	// required: true
	Username string `json:"username"`
	// Password пароль пользователя
	// required: true
	Password string `json:"password"`
	// DeviceID идентификатор устройства
	// required: true
	DeviceID string `json:"device_id"`
}

// AuthHTTPHandler обрабатывает HTTP-запросы, связанные с авторизацией.
type AuthHTTPHandler struct {
	svc               Authenticator
	usernameValidator func(username string) error
	passwordValidator func(password string) error
}

// NewAuthHTTPHandler создает новый HTTP-хендлер для авторизации.
func NewAuthHTTPHandler(
	svc Authenticator,
	usernameValidator func(username string) error,
	passwordValidator func(password string) error,
) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		svc:               svc,
		usernameValidator: usernameValidator,
		passwordValidator: passwordValidator,
	}
}

// Register обрабатывает регистрацию нового пользователя и устройства.
//
// swagger:route POST /register auth registerUser
//
// Регистрация нового пользователя и устройства.
//
// Consumes:
// - application/json
//
// Produces:
// - text/plain
//
// Parameters:
//   - name: body
//     in: body
//     description: Данные для регистрации
//     required: true
//     schema:
//     "$ref": "#/definitions/AuthRequest"
//
// Responses:
//
//	200: OK
//	400: Bad Request
//	409: Conflict
//	500: Internal Server Error
func (h *AuthHTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.usernameValidator != nil {
		if err := h.usernameValidator(req.Username); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if h.passwordValidator != nil {
		if err := h.passwordValidator(req.Password); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	privBytes, token, err := h.svc.Register(r.Context(), req.Username, req.Password, req.DeviceID)
	if err != nil {
		if errors.Is(err, services.ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
	w.Write(privBytes)
}
