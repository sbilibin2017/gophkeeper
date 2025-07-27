package http

import (
	"context"
	"encoding/json"
	"net/http"
)

// Registerer defines the interface for user registration.
type Registerer interface {
	Register(ctx context.Context, username, password string) (*string, error)
}

// Loginer defines the interface for user authentication (login).
type Loginer interface {
	Login(ctx context.Context, username, password string) (*string, error)
}

// User represents a user account in the system.
type AuthRequest struct {
	Username string `json:"username"` // Username is the unique identifier for the user.
	Password string `json:"password"` // PasswordHash is the hashed password.

}

// NewRegisterHandler returns an HTTP handler for user registration.
func NewRegisterHandler(reg Registerer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		token, err := reg.Register(ctx, req.Username, req.Password)
		if err != nil {
			http.Error(w, "registration failed", http.StatusInternalServerError)
			return
		}

		if token != nil {
			w.Header().Set("Authorization", "Bearer "+*token)
		}

		w.WriteHeader(http.StatusOK)
	}
}

// NewLoginHandler returns an HTTP handler for user login.
func NewLoginHandler(login Loginer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		token, err := login.Login(ctx, req.Username, req.Password)
		if err != nil {
			http.Error(w, "login failed", http.StatusUnauthorized)
			return
		}

		if token != nil {
			w.Header().Set("Authorization", "Bearer "+*token)
		}

		w.WriteHeader(http.StatusOK)
	}
}
