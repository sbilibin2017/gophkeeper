package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretTokenDecoder интерфейс для декодирования токена из HTTP-запроса.
type SecretTokenDecoder interface {
	GetFromRequest(req *http.Request) (string, error)
	Parse(tokenString string) (*models.Claims, error)
}

// SecretWriter интерфейс для сохранения секрета.
type SecretWriter interface {
	Save(ctx context.Context, secret *models.SecretDB) error
}

// SecretReader интерфейс для чтения секрета.
type SecretReader interface {
	Get(ctx context.Context, userID, secretID string) (*models.SecretDB, error)
	List(ctx context.Context, userID string) ([]*models.SecretDB, error)
}

// @Summary      Сохранение нового секрета
// @Description  Сохраняет новый секрет пользователя
// @Tags         secret
// @Accept       json
// @Produce      json
// @Param        secret body models.SecretRequest true "Данные секрета для сохранения"
// @Success      200 {string} string "Секрет успешно сохранен"
// @Failure      400 {string} string "Неверный запрос"
// @Failure      401 {string} string "Неавторизованный доступ"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /secret/save [post]
func NewSecretSaveHTTPHandler(
	tokenDecoder SecretTokenDecoder,
	secretWriter SecretWriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tokenString, err := tokenDecoder.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		claims, err := tokenDecoder.Parse(tokenString)
		if err != nil || claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req models.SecretRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		encryptedPayload, err := base64.StdEncoding.DecodeString(req.EncryptedPayload)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		nonce, err := base64.StdEncoding.DecodeString(req.Nonce)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		secret := &models.SecretDB{
			SecretID:         uuid.New().String(),
			UserID:           claims.UserID,
			SecretName:       req.SecretName,
			SecretType:       req.SecretType,
			EncryptedPayload: string(encryptedPayload),
			Nonce:            string(nonce),
			Meta:             req.Meta,
		}

		if err := secretWriter.Save(ctx, secret); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary      Получение секрета по ID
// @Description  Возвращает данные секрета пользователя по ID секрета
// @Tags         secret
// @Accept       json
// @Produce      json
// @Param        secret-id path string true "ID секрета"
// @Success      200 {object} models.SecretResponse
// @Failure      400 {string} string "Неверный запрос или отсутствует ID секрета"
// @Failure      401 {string} string "Неавторизованный доступ"
// @Failure      404 {string} string "Секрет не найден"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /secret/get/{secret-id} [get]
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

		secret, err := secretReader.Get(ctx, claims.UserID, secretID)
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
			EncryptedPayload: base64.StdEncoding.EncodeToString([]byte(secret.EncryptedPayload)),
			Nonce:            base64.StdEncoding.EncodeToString([]byte(secret.Nonce)),
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
// @Tags         secret
// @Accept       json
// @Produce      json
// @Success      200 {array} models.SecretResponse
// @Failure      400 {string} string "Неверный запрос"
// @Failure      401 {string} string "Неавторизованный доступ"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /secret/list [get]
func NewSecretListHTTPHandler(
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

		claims, err := tokenDecoder.Parse(tokenString)
		if err != nil || claims == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		secrets, err := secretReader.List(ctx, claims.UserID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := make([]models.SecretResponse, 0, len(secrets))
		for _, secret := range secrets {

			resp = append(resp, models.SecretResponse{
				SecretID:         secret.SecretID,
				UserID:           secret.UserID,
				SecretName:       secret.SecretName,
				SecretType:       secret.SecretType,
				EncryptedPayload: base64.StdEncoding.EncodeToString([]byte(secret.EncryptedPayload)),
				Nonce:            base64.StdEncoding.EncodeToString([]byte(secret.Nonce)),
				Meta:             secret.Meta,
				CreatedAt:        secret.CreatedAt,
				UpdatedAt:        secret.UpdatedAt,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
