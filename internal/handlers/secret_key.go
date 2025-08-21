package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretKeyTokenDecoder интерфейс для декодирования токена из HTTP-запроса.
type SecretKeyTokenDecoder interface {
	GetFromRequest(req *http.Request) (string, error)
	Parse(tokenString string) (*models.Claims, error)
}

// SecretKeyWriter интерфейс для сохранения секретного ключа.
type SecretKeyWriter interface {
	Save(ctx context.Context, secretKey *models.SecretKeyDB) error
}

// SecretKeyGetter интерфейс для получения секретного ключа.
type SecretKeyGetter interface {
	Get(ctx context.Context, secretID, deviceID string) (*models.SecretKeyDB, error)
}

// @Summary      Сохранение нового секретного ключа
// @Description  Сохраняет новый секретный ключ пользователя
// @Tags         secret-key
// @Accept       json
// @Produce      json
// @Param        secretKey body models.SecretKeyRequest true "Данные секретного ключа для сохранения"
// @Success      200 "Секретный ключ успешно сохранен"
// @Failure      400 "Неверный токен или некорректный запрос"
// @Failure      401 "Неавторизованный доступ"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /secret-key/save [post]
func NewSecretKeySaveHTTPHandler(
	tokenDecoder SecretKeyTokenDecoder,
	secretKeyWriter SecretKeyWriter,
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

		var req models.SecretKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		secretKey := &models.SecretKeyDB{
			SecretID:        req.SecretID,
			DeviceID:        claims.DeviceID,
			EncryptedAESKey: req.EncryptedAESKey,
		}

		if err := secretKeyWriter.Save(ctx, secretKey); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary      Получение информации о секретном ключе
// @Description  Извлекает секретный ключ по secret-id из URL и возвращает данные ключа
// @Tags         secret-key
// @Accept       json
// @Produce      json
// @Param        secret-id path string true "ID секрета"
// @Success      200 {object} models.SecretKeyResponse "Информация о секретном ключе"
// @Failure      400 "Неверный токен или некорректный запрос"
// @Failure      401 "Неавторизованный доступ"
// @Failure      404 "Секретный ключ не найден"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /secret-key/get/{secret-id} [get]
func NewSecretKeyGetHTTPHandler(
	tokenDecoder SecretKeyTokenDecoder,
	secretKeyGetter SecretKeyGetter,
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

		secretID := chi.URLParam(r, "secret-id")
		if secretID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		secretKey, err := secretKeyGetter.Get(ctx, secretID, claims.DeviceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if secretKey == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp := models.SecretKeyResponse{
			SecretKeyID:     secretKey.SecretKeyID,
			SecretID:        secretKey.SecretID,
			DeviceID:        secretKey.DeviceID,
			EncryptedAESKey: secretKey.EncryptedAESKey,
			CreatedAt:       secretKey.CreatedAt,
			UpdatedAt:       secretKey.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
