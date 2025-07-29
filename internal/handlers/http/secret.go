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

// NewSecretAddHandler returns an HTTP handler for adding a new secret.
//
// It expects a POST request with a JSON body matching models.SecretSaveRequest.
// Requires a Bearer token in the Authorization header.
// On success, responds with HTTP 200 OK.
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

		var req struct {
			SecretName string `json:"secret_name"`
			SecretType string `json:"secret_type"`
			Ciphertext []byte `json:"ciphertext"`
			AESKeyEnc  []byte `json:"aes_key_enc"`
		}
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

// NewSecretGetHandler returns an HTTP handler to retrieve a specific secret.
//
// Expects the secret type and name in the URL path, e.g., /secrets/{secret_type}/{secret_name}.
// Requires a Bearer token in the Authorization header.
// On success, responds with the secret as JSON.
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

// NewSecretListHandler returns an HTTP handler to list all secrets for an authenticated user.
//
// Requires a Bearer token in the Authorization header.
// On success, responds with a JSON array of all user's secrets.
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
