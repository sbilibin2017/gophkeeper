package handlers

import (
	"context"
	"encoding/json"
	"net/http"

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

// @Summary      Получение информации об устройстве
// @Description  Извлекает устройство и возвращает данные устройства
// @Tags         devices
// @Accept       json
// @Produce      json
// @Success      200 {object} handlers.DeviceResponse "Информация об устройстве"
// @Failure      400 "Неверный токен или запрос"
// @Failure      401 "Неавторизованный доступ"
// @Failure      404 "Устройство не найдено"
// @Failure      500 "Внутренняя ошибка сервера"
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
		resp := models.DeviceResponse{
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
