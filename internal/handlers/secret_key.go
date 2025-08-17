package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretKeyTokenDecoder интерфейс для декодирования токена из HTTP-запроса.
type SecretKeyTokenDecoder interface {
	// GetFromRequest извлекает токен из запроса
	GetFromRequest(req *http.Request) (string, error)
	// Parse парсит токен и возвращает secretID и deviceID
	Parse(tokenString string) (secretID string, deviceID string, err error)
}

// SecretKeyGetter интерфейс для получения секретного ключа из хранилища.
type SecretKeyWriter interface {
	Save(
		ctx context.Context,
		secretKeyID, secretID, deviceID string,
		encryptedAESKey []byte,
	) error
}

// SecretKeyGetter интерфейс для получения секретного ключа из хранилища.
type SecretKeyGetter interface {
	// Get возвращает секретный ключ по secretID и deviceID
	Get(ctx context.Context, secretID, deviceID string) (*models.SecretKeyDB, error)
}

// SecretKeyResponse описывает JSON-ответ с данными секретного ключа.
// swagger:model SecretKeyResponse
type SecretKeyResponse struct {
	// Уникальный идентификатор записи секретного ключа
	SecretKeyID string `json:"secret_key_id"`
	// Идентификатор секрета
	SecretID string `json:"secret_id"`
	// Идентификатор устройства
	DeviceID string `json:"device_id"`
	// AES ключ, зашифрованный публичным ключом устройства
	EncryptedAESKey []byte `json:"encrypted_aes_key"`
	// Дата создания записи
	CreatedAt time.Time `json:"created_at"`
	// Дата последнего обновления записи
	UpdatedAt time.Time `json:"updated_at"`
}

// @Summary      Сохранение нового секретного ключа
// @Description  Сохраняет новый секретный ключ пользователя, используя токен авторизации
// @Tags         secret-key
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer токен" default(Bearer <token>)
// @Param        secretKey body handlers.SecretKeyResponse true "Данные секретного ключа для сохранения"
// @Success      200 "Секретный ключ успешно сохранен"
// @Failure      400 "Неверный токен или запрос"
// @Failure      401 "Неавторизованный доступ"
// @Failure      500 "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /save-secret-key [post]
func NewSecretKeySaveHTTPHandler(
	tokenDecoder SecretKeyTokenDecoder,
	secretKeyWriter SecretKeyWriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context() // Контекст запроса

		// Извлекаем токен из запроса
		tokenString, err := tokenDecoder.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Парсим токен для получения secretID и deviceID
		secretID, deviceID, err := tokenDecoder.Parse(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Декодируем тело запроса в структуру SecretKeyResponse
		var req SecretKeyResponse
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Сохраняем секретный ключ
		err = secretKeyWriter.Save(ctx, req.SecretKeyID, secretID, deviceID, req.EncryptedAESKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Возвращаем успешный ответ
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary      Получение информации о секретном ключе
// @Description  Извлекает секретный ключ по токену из запроса, возвращает данные ключа
// @Tags         secret-key
// @Accept       json
// @Produce      json
// @Param        Authorization header string true "Bearer токен" default(Bearer <token>)
// @Success      200 {object} handlers.SecretKeyResponse "Информация о секретном ключе"
// @Failure      400 "Неверный токен или запрос"
// @Failure      401 "Неавторизованный доступ"
// @Failure      404 "Закодированный ключ секрета не найден"
// @Failure      500 "Внутренняя ошибка сервера"
// @Security     BearerAuth
// @Router       /get-secret-key [get]
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

		// Парсим токен
		secretID, deviceID, err := tokenDecoder.Parse(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Получаем закодирвоанный ключ секрета
		secretKey, err := secretKeyGetter.Get(ctx, secretID, deviceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if secretKey == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// Формируем JSON-ответ
		resp := SecretKeyResponse{
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
