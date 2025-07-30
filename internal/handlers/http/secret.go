package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretWriter defines interface to save secrets.
type SecretWriter interface {
	Save(ctx context.Context, username, secretName, secretType string, ciphertext, aesKeyEnc []byte) error
}

// SecretReader defines interface to read secrets.
type SecretReader interface {
	Get(ctx context.Context, username, secretType, secretName string) (*models.Secret, error)
	List(ctx context.Context, username string) ([]*models.Secret, error)
}

// JWTParser parses JWT token and returns username or error.
type JWTParser interface {
	Parse(token string) (username string, err error)
}

var ErrUnauthorized = errors.New("unauthorized")

// SecretSaveRequest represents request body for saving a secret.
// swagger:model SecretSaveRequest
type SecretSaveRequest struct {
	// Secret name
	// example: mysecret
	SecretName string `json:"secret_name" example:"mysecret"`
	// Secret type
	// example: password
	SecretType string `json:"secret_type" example:"password"`
	// Ciphertext bytes
	Ciphertext []byte `json:"ciphertext"`
	// Encrypted AES key
	AESKeyEnc []byte `json:"aes_key_enc"`
}

// SecretResponse represents secret data returned in responses.
// swagger:model SecretResponse
type SecretResponse struct {
	// Secret name
	SecretName string `json:"secret_name"`
	// Secret type
	SecretType string `json:"secret_type"`
	// Ciphertext bytes
	Ciphertext []byte `json:"ciphertext"`
	// Encrypted AES key
	AESKeyEnc []byte `json:"aes_key_enc"`
}

// NewSecretAddHandler returns an HTTP handler that saves a secret.
//
// @Summary Save a secret
// @Description Saves a secret for authenticated user
// @Tags secrets
// @Accept json
// @Produce json
// @Param secret body SecretSaveRequest true "Secret save request payload"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "invalid request body"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /secrets [post]
func NewSecretAddHandler(writer SecretWriter, parser JWTParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		username, err := parser.Parse(parts[1])
		if err != nil {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		var req SecretSaveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := writer.Save(ctx, username, req.SecretName, req.SecretType, req.Ciphertext, req.AESKeyEnc); err != nil {
			http.Error(w, "failed to save secret", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// NewSecretGetHandler returns an HTTP handler that retrieves a secret by type and name.
//
// @Summary Get a secret
// @Description Retrieves a secret for authenticated user by secret_type and secret_name
// @Tags secrets
// @Accept json
// @Produce json
// @Param secret_type path string true "Secret type"
// @Param secret_name path string true "Secret name"
// @Success 200 {object} SecretResponse
// @Failure 400 {string} string "missing parameters"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /secrets/{secret_type}/{secret_name} [get]
func NewSecretGetHandler(reader SecretReader, parser JWTParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		username, err := parser.Parse(parts[1])
		if err != nil {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		secretType := chi.URLParam(r, "secret_type")
		secretName := chi.URLParam(r, "secret_name")
		if secretType == "" || secretName == "" {
			http.Error(w, "missing secret_type or secret_name URL parameter", http.StatusBadRequest)
			return
		}

		secret, err := reader.Get(ctx, username, secretType, secretName)
		if err != nil {
			http.Error(w, "failed to get secret", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(secret); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

// NewSecretListHandler returns an HTTP handler that lists all secrets for a user.
//
// @Summary List all secrets
// @Description Lists all secrets for authenticated user
// @Tags secrets
// @Accept json
// @Produce json
// @Success 200 {array} SecretResponse
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "internal server error"
// @Router /secrets [get]
func NewSecretListHandler(reader SecretReader, parser JWTParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		parts := strings.Fields(authHeader)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		username, err := parser.Parse(parts[1])
		if err != nil {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		secrets, err := reader.List(ctx, username)
		if err != nil {
			http.Error(w, "failed to list secrets", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(secrets); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
