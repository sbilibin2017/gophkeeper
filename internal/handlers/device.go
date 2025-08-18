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
	Parse(tokenString string) (*models.Claims, error)
}

// DeviceGetter интерфейс для получения устройства из хранилища.
type DeviceGetter interface {
	// Get возвращает устройство по userID и deviceID
	Get(ctx context.Context, userID, deviceID string) (*models.DeviceDB, error)
}

// @Summary      Получение информации об устройстве
// @Description  Извлекает информацию о текущем устройстве по JWT токену из запроса
// @Tags         device
// @Accept       json
// @Produce      json
// @Success      200 {object} models.DeviceResponse "Информация об устройстве"
// @Failure      400 {string} string "Неверный токен или некорректный запрос"
// @Failure      401 {string} string "Неавторизованный доступ, неверный токен"
// @Failure      404 {string} string "Устройство не найдено"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /device/get [get]
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

		// Парсим токен и получаем claims
		claims, err := tokenDecoder.Parse(tokenString)
		if err != nil || claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Получаем устройство
		device, err := deviceGetter.Get(ctx, claims.UserID, claims.DeviceID)
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
