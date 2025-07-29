package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretWriter defines the interface for writing secrets to storage.
type SecretWriter interface {
	// Save stores a secret for the given user.
	Save(
		ctx context.Context,
		username string,
		secretName string,
		secretType string,
		ciphertext []byte,
		aesKeyEnc []byte,
	) error
}

// SecretReader defines the interface for reading secrets from storage.
type SecretReader interface {
	// Get retrieves a secret by type and name for the given user.
	Get(
		ctx context.Context,
		username string,
		secretType string,
		secretName string,
	) (*models.Secret, error)

	// List returns all secrets for the given user.
	List(
		ctx context.Context,
		username string,
	) ([]*models.Secret, error)
}

// JWTParser defines the interface for parsing JWT tokens.
type JWTParser interface {
	// Parse validates the token and returns the associated username.
	Parse(token string) (username string, err error)
}

// ErrUnauthorized represents an unauthorized access error due to missing or invalid token.
var ErrUnauthorized = errors.New("unauthorized")

// SecretSaveRequest is the expected request body for adding a secret.
// swagger:model SecretSaveRequest
type SecretSaveRequest struct {
	// SecretName is the unique name of the secret.
	// example: my-bank-password
	SecretName string `json:"secret_name"`
	// SecretType represents the type/category of the secret.
	// example: password
	SecretType string `json:"secret_type"`
	// Ciphertext is the encrypted secret data, base64 encoded.
	// example: SGVsbG8sIHNlY3JldCE=
	Ciphertext []byte `json:"ciphertext"`
	// AESKeyEnc is the encrypted AES key, base64 encoded.
	// example: U29tZUVuY3J5cHRlZEtleQ==
	AESKeyEnc []byte `json:"aes_key_enc"`
}

// SecretResponse models the secret data returned in responses.
// swagger:model SecretResponse
type SecretResponse struct {
	// SecretName is the unique name of the secret.
	SecretName string `json:"secret_name"`
	// SecretType represents the type/category of the secret.
	SecretType string `json:"secret_type"`
	// Ciphertext is the encrypted secret data, base64 encoded.
	Ciphertext []byte `json:"ciphertext"`
	// AESKeyEnc is the encrypted AES key, base64 encoded.
	AESKeyEnc []byte `json:"aes_key_enc"`
}

// @Summary Add a new secret
// @Description Adds a new secret for the authenticated user.
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param secret body SecretSaveRequest true "Secret data to save"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "invalid request body"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "failed to save secret"
// @Router /secrets [post]
func NewSecretAddHandler(
	secretWriter SecretWriter,
	jwtParser JWTParser,
) http.HandlerFunc {
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
		username, err := jwtParser.Parse(parts[1])
		if err != nil {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		var req SecretSaveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := secretWriter.Save(ctx, username, req.SecretName, req.SecretType, req.Ciphertext, req.AESKeyEnc); err != nil {
			http.Error(w, "failed to save secret", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Get a secret by type and name
// @Description Retrieves a secret for the authenticated user by secret type and name.
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param secret_type path string true "Secret type" example(password)
// @Param secret_name path string true "Secret name" example(my-bank-password)
// @Success 200 {object} SecretResponse
// @Failure 400 {string} string "missing secret_type or secret_name URL parameter"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "failed to get secret"
// @Router /secrets/{secret_type}/{secret_name} [get]
func NewSecretGetHandler(
	secretReader SecretReader,
	jwtParser JWTParser,
) http.HandlerFunc {
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
		username, err := jwtParser.Parse(parts[1])
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

		secret, err := secretReader.Get(ctx, username, secretType, secretName)
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

// @Summary List all secrets for a user
// @Description Returns all secrets belonging to the authenticated user.
// @Tags secrets
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} SecretResponse
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "failed to list secrets"
// @Router /secrets [get]
func NewSecretListHandler(
	secretReader SecretReader,
	jwtParser JWTParser,
) http.HandlerFunc {
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
		username, err := jwtParser.Parse(parts[1])
		if err != nil {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		secrets, err := secretReader.List(ctx, username)
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
