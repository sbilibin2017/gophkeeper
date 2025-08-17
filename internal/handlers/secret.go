package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretTokenDecoder интерфейс для декодирования токена из HTTP-запроса.
type SecretTokenDecoder interface {
	// GetFromRequest извлекает токен из HTTP-запроса.
	//
	// Параметры:
	//   req - указатель на HTTP-запрос
	//
	// Возвращает:
	//   токен в виде строки и ошибку в случае неудачи
	GetFromRequest(req *http.Request) (string, error)

	// Parse парсит токен и возвращает идентификаторы секрета и устройства.
	//
	// Параметры:
	//   tokenString - токен в виде строки
	//
	// Возвращает:
	//   secretID - идентификатор секрета
	//   deviceID - идентификатор устройства
	//   err - ошибка в случае неудачи
	Parse(tokenString string) (userID string, deviceID string, err error)
}

// SecretWriter интерфейс для сохранения секрета.
type SecretWriter interface {
	// Save сохраняет новый секрет в хранилище.
	//
	// Параметры:
	//   ctx - контекст запроса
	//   secretID - уникальный идентификатор секрета
	//   userID - идентификатор пользователя
	//   secretName - название секрета
	//   secretType - тип секрета
	//   encryptedPayload - зашифрованные данные секрета
	//   nonce - nonce для шифрования
	//   meta - метаданные секрета
	//
	// Возвращает:
	//   ошибку в случае неудачи
	Save(
		ctx context.Context,
		secretID, userID, secretName, secretType string,
		encryptedPayload, nonce []byte,
		meta string,
	) error
}

// SecretReader интерфейс для чтения секрета.
type SecretReader interface {
	// Get возвращает секрет по имени для указанного пользователя.
	//
	// Параметры:
	//   ctx - контекст запроса
	//   userID - идентификатор пользователя
	//   secretName - имя секрета
	//
	// Возвращает:
	//   указатель на SecretDB и ошибку в случае неудачи
	Get(ctx context.Context, userID, secretID string) (*models.SecretDB, error)

	// List возвращает список всех секретов указанного пользователя.
	//
	// Параметры:
	//   ctx - контекст запроса
	//   userID - идентификатор пользователя
	//
	// Возвращает:
	//   срез указателей на SecretDB и ошибку в случае неудачи
	List(ctx context.Context, userID string) ([]*models.SecretDB, error)
}

// @Summary      Сохранение нового секрета
// @Description  Сохраняет новый секрет пользователя
// @Tags         secrets
// @Accept       json
// @Produce      json
// @Param        secret body handlers.SecretRequest true "Данные секрета для сохранения"
// @Success      200 "Секрет успешно сохранен"
// @Failure      400 "Неверный запрос"
// @Failure      401 "Неавторизованный доступ"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /save-secret [post]
func NewSecretSaveHTTPHandler(
	tokenDecoder SecretTokenDecoder,
	secretWriter SecretWriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context() // Получаем контекст запроса для возможной отмены/таймаута

		// Извлекаем токен из запроса
		tokenString, err := tokenDecoder.GetFromRequest(r)
		if err != nil {
			// Если токен отсутствует или некорректный, возвращаем 400 Bad Request
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Парсим токен для получения идентификатора секрета и устройства
		secretID, _, err := tokenDecoder.Parse(tokenString)
		if err != nil {
			// Если токен недействителен, возвращаем 401 Unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Декодируем тело запроса в структуру SecretResponse
		var req models.SecretRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Если тело запроса невалидное, возвращаем 400 Bad Request
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Сохраняем секрет в хранилище
		err = secretWriter.Save(ctx, secretID, req.UserID, req.SecretName, req.SecretType, req.EncryptedPayload, req.Nonce, req.Meta)
		if err != nil {
			// Если возникла внутренняя ошибка при сохранении, возвращаем 500 Internal Server Error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Возвращаем успешный ответ
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary      Получение секрета по имени
// @Description  Возвращает данные секрета пользователя по имени секрета
// @Tags         secrets
// @Accept       json
// @Produce      json
// @Param        secret_name query string true "Имя секрета"
// @Success      200 {object} handlers.SecretResponse "Информация о секрете"
// @Failure      400 "Неверный запрос или отсутствует имя секрета"
// @Failure      401 "Неавторизованный доступ"
// @Failure      404 "Секрет не найден"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /get-secret [get]
func NewSecretGetHTTPHandler(
	tokenDecoder SecretTokenDecoder,
	secretReader SecretReader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString, err := tokenDecoder.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userID, _, err := tokenDecoder.Parse(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		secretID := chi.URLParam(r, "secret-id")
		if secretID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		secret, err := secretReader.Get(ctx, userID, secretID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if secret == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp := models.SecretResponse{
			SecretID:         secret.SecretID,
			UserID:           secret.UserID,
			SecretName:       secret.SecretName,
			SecretType:       secret.SecretType,
			EncryptedPayload: secret.EncryptedPayload,
			Nonce:            secret.Nonce,
			Meta:             secret.Meta,
			CreatedAt:        secret.CreatedAt,
			UpdatedAt:        secret.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// @Summary      Получение списка всех секретов пользователя
// @Description  Возвращает список всех секретов текущего пользователя
// @Tags         secrets
// @Accept       json
// @Produce      json
// @Success      200 {array} handlers.SecretResponse "Список секретов пользователя"
// @Failure      400 "Неверный запрос"
// @Failure      401 "Неавторизованный доступ"
// @Failure      500 "Внутренняя ошибка сервера"
// @Router       /list-secrets [get]
func NewSecretListHTTPHandler(
	tokenDecoder SecretTokenDecoder,
	secretReader SecretReader,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context() // Получаем контекст запроса

		// Извлекаем токен из запроса
		tokenString, err := tokenDecoder.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Парсим токен для получения идентификатора пользователя
		_, userID, err := tokenDecoder.Parse(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Получаем список всех секретов пользователя
		secrets, err := secretReader.List(ctx, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Преобразуем секреты в формат ответа
		var resp []models.SecretResponse
		for _, secret := range secrets {
			resp = append(resp, models.SecretResponse{
				SecretID:         secret.SecretID,
				UserID:           secret.UserID,
				SecretName:       secret.SecretName,
				SecretType:       secret.SecretType,
				EncryptedPayload: secret.EncryptedPayload,
				Nonce:            secret.Nonce,
				Meta:             secret.Meta,
				CreatedAt:        secret.CreatedAt,
				UpdatedAt:        secret.UpdatedAt,
			})
		}

		// Устанавливаем заголовок и кодируем ответ в JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
