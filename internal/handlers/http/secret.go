package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sbilibin2017/gophkeeper/internal/models"
)

// SecretWriter defines write operations for secrets.
type SecretWriter interface {
	Save(
		ctx context.Context,
		secretOwner string,
		secretName string,
		secretType string,
		ciphertext []byte,
		aesKeyEnc []byte,
	) error
}

// SecretReader defines read operations for secrets.
type SecretReader interface {
	Get(ctx context.Context, secretOwner, secretType, secretName string) (*models.Secret, error)
	List(ctx context.Context, secretOwner string) ([]*models.Secret, error)
}

// JWTParser defines the interface for parsing JWT tokens.
type JWTParser interface {
	Parse(tokenStr string) (string, error)
}

// NewSecretAddHandler returns an HTTP handler for adding a new secret.
func NewSecretAddHandler(
	secretWriter SecretWriter,
	jwtParser JWTParser,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		parts := strings.Fields(authHeader)
		if authHeader == "" || len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid or missing authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		username, err := jwtParser.Parse(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		var req models.SecretSaveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		err = secretWriter.Save(ctx, username, req.SecretName, req.SecretType, req.Ciphertext, req.AESKeyEnc)
		if err != nil {
			http.Error(w, "failed to save secret", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// NewSecretGetHandler returns an HTTP handler to retrieve a secret by its type and name.
func NewSecretGetHandler(
	secretReader SecretReader,
	jwtParser JWTParser,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		parts := strings.Fields(authHeader)
		if authHeader == "" || len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid or missing authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		username, err := jwtParser.Parse(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
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

// NewSecretListHandler returns an HTTP handler to list all secrets for the authenticated user.
func NewSecretListHandler(
	secretReader SecretReader,
	jwtParser JWTParser,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		parts := strings.Fields(authHeader)
		if authHeader == "" || len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid or missing authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		username, err := jwtParser.Parse(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
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
