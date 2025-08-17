package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// TokenDecoder интерфейс для декодирования токена из HTTP-запроса.
type TokenDecoder interface {
	// GetFromRequest извлекает токен из запроса
	GetFromRequest(req *http.Request) (string, error)
	// Parse парсит токен и возвращает userID и deviceID
	Parse(tokenString string) (userID string, deviceID string, err error)
}

// DeviceGetter интерфейс для получения устройства из хранилища.
type DeviceGetter interface {
	// Get возвращает устройство по userID и deviceID
	Get(ctx context.Context, userID, deviceID string) (*models.DeviceDB, error)
}

// DeviceResponse описывает JSON-ответ с данными устройства.
// swagger:model DeviceResponse
type DeviceResponse struct {
	// Уникальный идентификатор устройства
	DeviceID string `json:"device_id"`
	// Идентификатор пользователя-владельца устройства
	UserID string `json:"user_id"`
	// Публичный ключ устройства
	PublicKey string `json:"public_key"`
	// Дата создания устройства
	CreatedAt time.Time `json:"created_at"`
	// Дата последнего обновления данных устройства
	UpdatedAt time.Time `json:"updated_at"`
}

// @Summary      Получение информации об устройстве
// @Description  Извлекает устройство по токену из запроса, возвращает данные устройства
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer токен" default(Bearer <token>)
// @Success      200 {object} handlers.DeviceResponse "Информация об устройстве"
// @Failure      400 "Неверный токен или запрос"
// @Failure      401 "Неавторизованный доступ"
// @Failure      404 "Устройство не найдено"
// @Failure      500 "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /get-device [get]
func NewDeviceGetHTTPHandler(
	tokenDecoder TokenDecoder,
	deviceGetter DeviceGetter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Извлекаем токен из запроса
		tokenString, err := tokenDecoder.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Парсим токен
		userID, deviceID, err := tokenDecoder.Parse(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Получаем устройство
		device, err := deviceGetter.Get(ctx, userID, deviceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if device == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Формируем JSON-ответ
		resp := DeviceResponse{
			DeviceID:  device.DeviceID,
			UserID:    device.UserID,
			PublicKey: device.PublicKey,
			CreatedAt: device.CreatedAt,
			UpdatedAt: device.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
